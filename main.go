package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type IMUData struct {
	Participant   string    `json:"participant"`
	Exercise      string    `json:"exercise"`
	Timestamp     time.Time `json:"timestamp"`
	RepID         int       `json:"rep_id"`
	IsRepActive   bool      `json:"is_rep_active"`
	Accelerometer []float64 `json:"accelerometer"`
	Gyroscope     []float64 `json:"gyroscope"`
	Magnetometer  []float64 `json:"magnetometer"`
}

var (
	influxClient influxdb2.Client
	org          string
	bucket       string
)

func main() {
	influxClient = influxdb2.NewClient(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("INFLUXDB_TOKEN"),
	)
	defer influxClient.Close()

	org = os.Getenv("INFLUXDB_ORG")
	bucket = os.Getenv("INFLUXDB_BUCKET")

	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/imu/csv", handleCSVData)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(router.Run(":" + port))
}

func createInfluxPoint(data IMUData) (*write.Point, error) {
	if len(data.Accelerometer) != 3 || len(data.Gyroscope) != 3 || len(data.Magnetometer) != 3 {
		return nil, fmt.Errorf("invalid sensor data format")
	}

	return influxdb2.NewPointWithMeasurement("exercise_data").
		AddTag("participant", data.Participant).
		AddTag("exercise", data.Exercise).
		AddTag("rep_id", strconv.Itoa(data.RepID)).
		AddField("rep_active", data.IsRepActive).
		AddField("accelerometer_x", data.Accelerometer[0]).
		AddField("accelerometer_y", data.Accelerometer[1]).
		AddField("accelerometer_z", data.Accelerometer[2]).
		AddField("gyroscope_x", data.Gyroscope[0]).
		AddField("gyroscope_y", data.Gyroscope[1]).
		AddField("gyroscope_z", data.Gyroscope[2]).
		AddField("magnetometer_x", data.Magnetometer[0]).
		AddField("magnetometer_y", data.Magnetometer[1]).
		AddField("magnetometer_z", data.Magnetometer[2]).
		SetTime(data.Timestamp), nil
}

func handleCSVData(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	r := csv.NewReader(bytes.NewReader(body))
	records, err := r.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data records provided"})
		return
	}

	writeAPI := influxClient.WriteAPI(org, bucket)
	errorCount := 0

	for i, record := range records[1:] {
		if len(record) < 14 {
			log.Printf("Line %d: insufficient columns", i+2)
			errorCount++
			continue
		}

		timestamp, err := time.Parse(time.RFC3339, record[2])
		if err != nil {
			log.Printf("Line %d: invalid timestamp: %v", i+2, err)
			errorCount++
			continue
		}

		repID, err := strconv.Atoi(record[3])
		if err != nil {
			log.Printf("Line %d: invalid rep_id: %v", i+2, err)
			errorCount++
			continue
		}

		isRepActive, err := strconv.ParseBool(record[4])
		if err != nil {
			log.Printf("Line %d: invalid is_rep_active: %v", i+2, err)
			errorCount++
			continue
		}

		sensorValues := make([]float64, 9)
		for j := 5; j < 14; j++ {
			val, err := strconv.ParseFloat(record[j], 64)
			if err != nil {
				log.Printf("Line %d: invalid sensor value at column %d: %v", i+2, j+1, err)
				errorCount++
				continue
			}
			sensorValues[j-5] = val
		}

		data := IMUData{
			Participant:   record[0],
			Exercise:      record[1],
			Timestamp:     timestamp,
			RepID:         repID,
			IsRepActive:   isRepActive,
			Gyroscope:     sensorValues[0:3],
			Accelerometer: sensorValues[3:6],
			Magnetometer:  sensorValues[6:9],
		}

		point, err := createInfluxPoint(data)
		if err != nil {
			log.Printf("Line %d: %v", i+2, err)
			errorCount++
			continue
		}

		writeAPI.WritePoint(point)
	}

	go func() {
		errorsCh := writeAPI.Errors()
		for err := range errorsCh {
			log.Printf("Write error: %s", err.Error())
			errorCount++
		}
	}()

	totalRecords := len(records) - 1
	if errorCount > 0 {
		c.JSON(http.StatusMultiStatus, gin.H{
			"message":    "CSV partially processed",
			"total":      totalRecords,
			"errors":     errorCount,
			"successful": totalRecords - errorCount,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "CSV successfully processed",
			"total":   totalRecords,
		})
	}
}