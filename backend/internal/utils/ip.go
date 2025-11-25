package utils

import (
	"log"
	"net"
)

func GetOutboundIP() string {
	// Se conecta a una dirección pública (Google DNS) para determinar la interfaz de salida,
	// pero sin enviar datos. Esto es seguro y rápido.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Println("Advertencia: No se pudo determinar la IP de salida. Usando localhost por defecto.")
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
