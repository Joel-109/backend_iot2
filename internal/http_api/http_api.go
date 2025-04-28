package httpapi

import (
	. "backend_iot2/internal/entities"
	"backend_iot2/internal/repository"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func StartWebServer(ctx context.Context, repo *repository.Queries, mqttClient *autopaho.ConnectionManager) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /sensors/{limit}", getSensorValues(ctx, repo))
	mux.HandleFunc("GET /sensor", getLastSensorValues(ctx, repo))
	mux.HandleFunc("GET /risk", getRisk(ctx, repo))
	mux.HandleFunc("GET /risks/{limit}", getRisks(ctx, repo))
	mux.HandleFunc("POST /config", postConfig(ctx, mqttClient))
	http.ListenAndServe(":8080", corsMiddleware(mux))
}

func getSensorValues(ctx context.Context, repo *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		limit, err := strconv.Atoi(r.PathValue("limit"))

		sensorValues, err := repo.GetSensorValues(ctx, int64(limit))

		if err != nil {
			http.Error(w, "Error en el GET", 400)
		}
		w.Header().Set("Content-Type", "application/json")

		data, err := json.Marshal(sensorValues)

		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func getLastSensorValues(ctx context.Context, repo *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorValues, err := repo.GetLastSensorValue(ctx)

		if err != nil {
			http.Error(w, "Error en el GET", 400)
		}
		w.Header().Set("Content-Type", "application/json")

		data, err := json.Marshal(sensorValues)

		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func getRisk(ctx context.Context, repo *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		risk, err := repo.GetLastRisk(ctx)

		if err != nil {
			http.Error(w, "Error en el GET", 400)
		}
		w.Header().Set("Content-Type", "application/json")

		data, err := json.Marshal(risk)

		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func getRisks(ctx context.Context, repo *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.PathValue("limit"))

		risk, err := repo.GetRisks(ctx, int64(limit))

		if err != nil {
			http.Error(w, "Error en el GET", 400)
		}
		w.Header().Set("Content-Type", "application/json")

		data, err := json.Marshal(risk)

		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func postConfig(ctx context.Context, mqttClient *autopaho.ConnectionManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var configValues ConfigValues
		err := json.NewDecoder(r.Body).Decode(&configValues)

		if err != nil {
			http.Error(w, "Error Decoding Config", 500)
		}

		mqttClient.Publish(
			ctx,
			&paho.Publish{
				QoS:     1,
				Topic:   "config/set",
				Payload: configValues.ToBytes(),
			},
		)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
