package main

import(
    "net/http"
    "log"
    "fmt"
    "github.com/mahimapatel13/ride-sharing-system/internal/config"
)

func main(){
    log.Printf("Bootstrapping sytem..")

    log.Printf("Loading .env ..")
    cfg := config.LoadEnv()

    // 
}