/*
 * Copyright (c) 2024 Contributors to the Eclipse Foundation
 *
 *  All rights reserved. This program and the accompanying materials
 *  are made available under the terms of the Eclipse Public License v2.0
 *  and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 *  and the Eclipse Distribution License is available at
 *    http://www.eclipse.org/org/documents/edl-v10.php.
 *
 *  SPDX-License-Identifier: EPL-2.0 OR BSD-3-Clause
 */

package main

import (
	. "backend_iot2/internal/entities"
	httpapi "backend_iot2/internal/http_api"
	"backend_iot2/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// Change this to something random if using a public test server

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error")
	}

	mqttBrokerUrl := os.Getenv("MQTT_BROKER")
	clientID := os.Getenv("CLIENT_ID")

	// App will run until cancelled by user (e.g. ctrl-c)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// We will connect to the Eclipse test server (note that you may see messages that other users publish)
	u, err := url.Parse(mqttBrokerUrl)
	if err != nil {
		panic(err)
	}

	conn, err := sql.Open("sqlite3", "./Values.db")

	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}

	defer conn.Close()

	repo := repository.New(conn)

	cliCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{u},
		KeepAlive:  20, // Keepalive message should be sent every 20 seconds
		// CleanStartOnInitialConnection defaults to false. Setting this to true will clear the session on the first connection.
		CleanStartOnInitialConnection: false,
		// SessionExpiryInterval - Seconds that a session will survive after disconnection.
		// It is important to set this because otherwise, any queued messages will be lost if the connection drops and
		// the server will not queue messages while it is down. The specific setting will depend upon your needs
		// (60 = 1 minute, 3600 = 1 hour, 86400 = one day, 0xFFFFFFFE = 136 years, 0xFFFFFFFF = don't expire)
		SessionExpiryInterval: 60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
			// Subscribing in the OnConnectionUp callback is recommended (ensures the subscription is reestablished if
			// the connection drops)
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: "sensors", QoS: 1},
					{Topic: "risk", QoS: 1},
					{Topic: "config", QoS: 1},
				},
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			}
			fmt.Println("mqtt subscription made")
		},
		OnConnectError: func(err error) { fmt.Printf("error whilst attempting connection: %s\n", err) },
		// eclipse/paho.golang/paho provides base mqtt functionality, the below config will be passed in for each connection
		ClientConfig: paho.ClientConfig{
			// If you are using QOS 1/2, then it's important to specify a client id (which must be unique)
			ClientID: clientID,
			// OnPublishReceived is a slice of functions that will be called when a message is received.
			// You can write the function(s) yourself or use the supplied Router
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					if pr.Packet.Topic == "sensors" {
						sensorValues := CreateSensorDataFromBytes(pr.Packet.Payload)
						repo.InsertSensorValues(ctx, repository.InsertSensorValuesParams{
							Temperature: float64(sensorValues.Temperature),
							Gas:         int64(sensorValues.Gas),
							Flame:       sensorValues.Flame,
						})
					}
					if pr.Packet.Topic == "risk" {
						risk := Risk{Risk: pr.Packet.Payload[0]}
						repo.InsertRisk(ctx, int64(risk.Risk))
					}
					if pr.Packet.Topic == "config" {
						configValues := CreateConfigFromBytes(pr.Packet.Payload)
						fmt.Printf("Sensors Message: %s \n", configValues)
					}
					return true, nil
				}},
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	c, err := autopaho.NewConnection(ctx, cliCfg) // starts process; will reconnect until context cancelled

	go func() {
		httpapi.StartWebServer(ctx, repo, c)
	}()

	if err != nil {
		panic(err)
	}
	// Wait for the connection to come up
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}

	fmt.Println("signal caught - exiting")
	<-c.Done() // Wait for clean shutdown (cancelling the context triggered the shutdown)

	fmt.Println("Ejecución Finalizada")
}
