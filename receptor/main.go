package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
  "github.com/labdao/receptor/models"
)

func main() {
  models.ConnectDatabase()

  r := gin.Default()

  r.GET("/_health", health)
  r.POST("/judge", judge)

  r.Run()
}

func health(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"status": "ok"})    
}

func judge(c *gin.Context) {

  // createa job row
  job := models.Job{}
  if err:=c.BindJSON(&job);err!=nil{
    c.AbortWithError(http.StatusBadRequest,err)
    return
  }

  models.DB.Create(&job)

  // the judge endpoint always returns status 200 to accept all jobs (for now)
  c.JSON(http.StatusOK, gin.H{})    
}
