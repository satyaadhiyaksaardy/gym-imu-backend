package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type IMUData struct {
	Participant  string    `json:"participant"`
	Exercise     string    `json:"exercise"`
	Timestamp    time.Time `json:"timestamp"`
	RepID        int       `json:"rep_id"`
	IsRepActive  bool      `json:"is_rep_active"`
	Accelerometer []float64 `json:"accelerometer"`
	Gyroscope    []float64 `json:"gyroscope"`
	Magnetometer []float64 `json:"magnetometer"`
}

var (
	influxClient influxdb2.Client
	org          string
	bucket       string
)

func main() {
	// Initialize InfluxDB client
	influxClient = influxdb2.NewClient(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("INFLUXDB_TOKEN"),
	)
	defer influxClient.Close()

	org = os.Getenv("INFLUXDB_ORG")
	bucket = os.Getenv("INFLUXDB_BUCKET")

	// Create Gin router
	router := gin.Default()
	router.Use(cors.Default())

	// API endpoint
	router.POST("/imu", handleIMUData)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(router.Run(":" + port))
}

func handleIMUData(c *gin.Context) {
	var data IMUData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate sensor data
	if len(data.Accelerometer) != 3 || len(data.Gyroscope) != 3 || len(data.Magnetometer) != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor data format"})
		return
	}

	// Write to InfluxDB
	writeAPI := influxClient.WriteAPIBlocking(org, bucket)
	p := influxdb2.NewPointWithMeasurement("exercise_data").
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
		SetTime(data.Timestamp)

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data stored successfully"})
}