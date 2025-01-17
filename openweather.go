package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const openWeatherAPIURL = "http://api.openweathermap.org/data/2.5/forecast?q=Moscow&appid=%s"

func (forecast *WatherForecast) updateOpenWeather() {
	var myClient = &http.Client{Timeout: 30 * time.Second}

	uri := fmt.Sprintf(openWeatherAPIURL, os.Getenv(openWeatherEnvVar))

	res, err := myClient.Get(uri)
	if err != nil {
		log.Printf("Get '%s': %v", uri, err)
		return
	}

	if res.StatusCode != 200 {
		log.Printf("Fetch weather error: %d %s", res.StatusCode, res.Status)
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	var jval OpenWeatherForecast
	if err := json.Unmarshal(body, &jval); err != nil {
		log.Printf("json.Unmarshal '%s': %v", body, err)
		return
	}

	if len(jval.List) == 0 {
		log.Printf("No forecast")
		return
	}

	forecast.mut.Lock()
	defer forecast.mut.Unlock()

	if jval.List[0].Clouds.All < 33 {
		forecast.CloudPrediction = 3
	} else if jval.List[0].Clouds.All < 66 {
		forecast.CloudPrediction = 2
	} else {
		forecast.CloudPrediction = 1
	}

	// ID mapping: https://openweathermap.org/weather-conditions
	for _, w := range jval.List[0].Weather {
		if w.ID >= 500 && w.ID < 600 {
			forecast.RainPrediction = 2
		} else {
			forecast.RainPrediction = 1
		}

		if w.ID >= 200 && w.ID < 300 {
			forecast.RainPrediction = 2
		}

		if w.ID >= 800 {
			if w.ID == 800 {
				forecast.CloudPrediction = 3
			} else if w.ID < 803 {
				forecast.CloudPrediction = 2
			} else {
				forecast.CloudPrediction = 1
			}
		}
	}

	forecast.updateTime = time.Now()
}

type OpenWeatherForecastItem struct {
	Dt int `json:"dt"`
	Main struct {
		Temp float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity int `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		ID int `json:"id"`
		Main string `json:"main"`
		Description string `json:"description"`
		Icon string `json:"icon"`
	} `json:"weather"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg float64 `json:"deg"`
	} `json:"wind"`
}

type OpenWeatherForecast struct {
	List []OpenWeatherForecastItem `json:"list"`
}
