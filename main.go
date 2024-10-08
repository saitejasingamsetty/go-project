package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var usersMap = make(map[string]Member)
var classesMap = make(map[string]Class)
var bookingsMap = make(map[string]Booking)

//var bookings = []&Booking

type Member struct {
	Name         string `json: "name" validate:"required"`
	MobileNumber string `json: "mobileNumber"`
	IsActive     bool   `json: isActive, default=true`
}

type Classes struct {
	ClassName string `json: "className"`
	FromDate  string `json: "fromDate"`
	ToDate    string `json: "toDate"`
	Capacity  int    `json: "capacity"`
}

type Booking struct {
	ClassName    string `json: "className"`
	MobileNumber string `json: "mobileNumber"`
	Name         string `json: "Name"`
	Date         string `json: "date"`
}

type Class struct {
	ClassName string `json: "classname"`
	Date      string `json: "date"`
	Capacity  int    `json: "capacity"`
	IsActive  bool   `json: isActive`
}

const (
	ISOFormat = "2006-01-02"
)

func healthcheck(request *gin.Context) {
	request.JSON(http.StatusOK, gin.H{"message": "Hi I am running"})
}

func createClasess(request *gin.Context) {

	var class Classes
	var existingDates []string
	var createdDates []string
	var skipMessage string
	var successMessage string

	err := request.ShouldBindBodyWithJSON(&class)
	if err != nil {
		request.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	fromDate, err := time.Parse(ISOFormat, class.FromDate)
	toDate, err := time.Parse(ISOFormat, class.ToDate)
	diffDays := toDate.Sub(fromDate) / (24 * time.Hour)
	classname := class.ClassName
	for day := 0; day <= int(diffDays); day++ {
		classDate := fromDate.AddDate(0, 0, day).Format(ISOFormat)
		_id := fmt.Sprintf("%v#%v", classname, classDate)
		_, IsClassExists := classesMap[_id]

		if IsClassExists {
			existingDates = append(existingDates, classDate)
			continue
		}

		classData := Class{
			ClassName: class.ClassName,
			Date:      classDate,
			Capacity:  class.Capacity,
			IsActive:  true,
		}
		classesMap[_id] = classData
		createdDates = append(createdDates, classDate)
	}

	if len(existingDates) > 0 {
		skipMessage = "classes are already exists on these dates"

	}

	if len(createdDates) > 0 {
		successMessage = "classes are created for these dates"
	}
	request.JSON(http.StatusMultiStatus,
		gin.H{"success_status": gin.H{"status": http.StatusCreated, "created_dates": createdDates, "message": successMessage},
			"failure_status": gin.H{"status": http.StatusConflict, "skipped_dates": existingDates, "message": skipMessage}})
}

func createMember(request *gin.Context) {

	var member Member
	err := request.ShouldBindBodyWithJSON(&member)
	if err != nil {
		request.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	member.IsActive = true

	mobileNumber := member.MobileNumber
	_, isMemberExists := usersMap[mobileNumber]

	if isMemberExists {
		request.JSON(http.StatusConflict, gin.H{"message": "Member is already exists with the same mobile number"})
		return
	}
	usersMap[mobileNumber] = member
	fmt.Println(usersMap)

	request.JSON(http.StatusCreated, gin.H{"message": "Successfully Registered Member",
		"Name":          member.Name,
		"Mobile Number": member.MobileNumber,
		"status":        member.IsActive})
}

func bookClass(request *gin.Context) {
	var booking Booking
	err := request.ShouldBindBodyWithJSON(&booking)

	if err != nil {

		request.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
	mobileNumber := booking.MobileNumber
	date := booking.Date
	className := booking.ClassName
	booking_id := fmt.Sprintf("%v#%v#%v", mobileNumber, className, date)
	class_id := fmt.Sprintf("%v#%v", className, date)

	_, isClassExists := classesMap[class_id]

	if !isClassExists {
		request.JSON(http.StatusNotFound, gin.H{"message": "class not exist"})
		return
	}

	_, isMemberExists := usersMap[mobileNumber]
	if !isMemberExists {
		request.JSON(http.StatusNotFound, gin.H{"message": "Member Not exists"})
		return
	}

	_, isBookingExists := bookingsMap[booking_id]

	if isBookingExists {
		request.JSON(http.StatusConflict, gin.H{"messsage": "You have already booked for the class"})
		return
	}

	bookingsMap[booking_id] = booking

	request.JSON(http.StatusCreated, gin.H{"message": "You have successfully booked for the class"})

}

func main() {

	route := gin.Default()
	route.GET("/healthCheck", healthcheck)
	route.POST("/createMember", createMember)
	route.POST("/classes", createClasess)
	route.POST("/bookings", bookClass)

	route.Run()
}
