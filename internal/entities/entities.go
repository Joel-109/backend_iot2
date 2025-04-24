package entities

import (
	"encoding/binary"
)

type SensorsData struct {
	Temperature float32 `json:temperature`
	Gas         uint16  `json:gas`
	Flame       bool    `json:flame`
}

type ConfigValues struct {
	TempTreshold         float32
	GasTreshold          uint16
	AlarmsEnabled        bool
	DataPointingInterval uint8
}

type Risk struct {
	Risk uint8
}

func (r *Risk) isHigh() bool {
	if r.Risk == 2 {
		return true
	}
	return false
}

func (r *Risk) isModerate() bool {
	if r.Risk == 1 {
		return true
	}
	return false
}
func (r *Risk) isLow() bool {
	if r.Risk == 0 {
		return true
	}
	return false
}

func CreateSensorDataFromBytes(array []byte) SensorsData {

	temperature_int := binary.LittleEndian.Uint16([]byte{array[0], array[1]})
	temperature := float32(temperature_int) / 100

	gas := binary.LittleEndian.Uint16([]byte{array[2], array[3]})
	var flame bool

	if array[4] == 0 {
		flame = false
	} else {
		flame = true
	}

	return SensorsData{temperature, gas, flame}
}

func CreateConfigFromBytes(array []byte) ConfigValues {
	temperatureIntTreshold := binary.LittleEndian.Uint16([]byte{array[0], array[1]})
	temperatureTreshold := float32(temperatureIntTreshold) / 100

	gas := binary.LittleEndian.Uint16([]byte{array[2], array[3]})

	var alarmsEnabled bool

	if array[4] == 0 {
		alarmsEnabled = false
	} else {
		alarmsEnabled = true
	}

	return ConfigValues{temperatureTreshold, gas, alarmsEnabled, array[5]}
}

func (c *ConfigValues) ToBytes() []byte {
	// Suponemos que TempTreshold es un float, y queremos escalarlo a centésimas
	temperatureInt := uint16(c.TempTreshold * 100)
	gasInt := uint16(c.GasTreshold)

	// Crear slice con capacidad para 5 bytes: 2 para temperatura, 2 para gas, 1 para flama
	buf := make([]byte, 6)

	// Escribir valores en orden Little Endian
	binary.LittleEndian.PutUint16(buf[0:2], temperatureInt)
	binary.LittleEndian.PutUint16(buf[2:4], gasInt)

	// Suponiendo que `c.Flame` es bool; si no, ajustalo según el tipo real

	if c.AlarmsEnabled {
		buf[4] = 1
	} else {
		buf[4] = 0
	}

	buf[5] = c.DataPointingInterval

	return buf
}
