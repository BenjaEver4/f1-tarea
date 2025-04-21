package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("===== MENÚ PRINCIPAL =====")
		fmt.Println("1. Ver todos los corredores")
		fmt.Println("2. Ver detalle de un corredor")
		fmt.Println("3. Ver todas las carreras")
		fmt.Println("4. Ver detalle de una carrera")
		fmt.Println("5. Ver resumen de temporada")
		fmt.Println("6. Salir")
		fmt.Print("Selecciona una opción: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		option, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ Opción inválida. Intenta nuevamente.\n")
			continue
		}

		switch option {
		case 1:
			fmt.Println("\n🔎 Obteniendo lista de corredores...")
		
			resp, err := http.Get("http://localhost:8080/api/corredor")
			if err != nil {
				fmt.Println(" Error al conectar con el servidor:", err)
				break
			}
			defer resp.Body.Close()
		
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(" Error al leer la respuesta:", err)
				break
			}
		
			var corredores []struct {
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				DriverNumber int    `json:"driver_number"`
				TeamName     string `json:"team_name"`
				CountryCode  string `json:"country_code"`
			}
		
			if err := json.Unmarshal(body, &corredores); err != nil {
				fmt.Println("❌ Error al parsear los datos:", err)
				break
			}
		
			// Imprimir tabla
			fmt.Println("\n| # | Nombre   | Apellido     | N Piloto | Equipo            | País |")
			fmt.Println("---------------------------------------------------------------------")
			for i, c := range corredores {
				fmt.Printf("| %-2d| %-8s | %-12s | %-8d | %-17s | %-4s |\n",
					i+1, c.FirstName, c.LastName, c.DriverNumber, c.TeamName, c.CountryCode)
			}
			fmt.Println()
		case 2:
			fmt.Print("\nIngrese el numero del piloto: ")
			idInput, _ := reader.ReadString('\n')
			idInput = strings.TrimSpace(idInput)
		
			url := fmt.Sprintf("http://localhost:8080/api/corredor/detalle/%s", idInput)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println(" Error al conectar con el servidor:", err)
				break
			}
			defer resp.Body.Close()
		
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(" Error al leer la respuesta:", err)
				break
			}
		
			var detalle struct {
				DriverID           string `json:"driver_id"`
				PerformanceSummary struct {
					Wins     int     `json:"wins"`
					Top3Fin  int     `json:"top_3_finishes"`
					MaxSpeed float64 `json:"max_speed"`
				} `json:"performance_summary"`
				RaceResults []struct {
					SessionKey       int     `json:"session_key"`
					Race             string  `json:"race"`
					CircuitShortName string  `json:"circuit_short_name"`
					Position         int     `json:"position"`
					FastestLap       bool    `json:"fastest_lap"`
					MaxSpeed         float64 `json:"max_speed"`
					BestLapDuration  float64 `json:"best_lap_duration"`
				} `json:"race_results"`
			}
		
			if err := json.Unmarshal(body, &detalle); err != nil {
				fmt.Println(" Error al parsear los datos:", err)
				break
			}
		
			fmt.Println("\n====================================================================================================")
			fmt.Println("| # | Carrera                | Pos Final | Vuelta rápida | Velocidad max | Menor tiempo vuelta     |")
			fmt.Println("====================================================================================================")
		
			for i, r := range detalle.RaceResults {
				vueltaRapida := "No"
				if r.FastestLap {
					vueltaRapida = "Sí"
				}
				fmt.Printf("| %-2d| %-23s | %-9d | %-13s | %-14.0f | %-21.3f |\n",
					i+1, r.Race, r.Position, vueltaRapida, r.MaxSpeed, r.BestLapDuration)
			}
			fmt.Println("====================================================================================================")
		
			fmt.Println("\n============================")
			fmt.Println("| Resumen del piloto       |")
			fmt.Println("============================")
			fmt.Printf("| Carreras ganadas         | %-4d |\n", detalle.PerformanceSummary.Wins)
			fmt.Printf("| Veces en el top 3        | %-4d |\n", detalle.PerformanceSummary.Top3Fin)
			fmt.Printf("| Velocidad máxima alcanzada | %.0f km/h |\n", detalle.PerformanceSummary.MaxSpeed)
			fmt.Println("============================")
			fmt.Println("\n===== MENÚ PRINCIPAL =====")
		case 3:
			fmt.Println("[3] Ver todas las carreras\n")
		
			resp, err := http.Get("http://localhost:8080/api/carrera")
			if err != nil {
				fmt.Println(" Error al hacer la solicitud:", err)
				break
			}
			defer resp.Body.Close()
		
			var carreras []struct {
				SessionKey        int    `json:"session_key"`
				CountryName       string `json:"country_name"`
				DateStart         string `json:"date_start"`
				Year              int    `json:"year"`
				CircuitShortName  string `json:"circuit_short_name"`
			}
		
			if err := json.NewDecoder(resp.Body).Decode(&carreras); err != nil {
				fmt.Println("❌ Error al leer la respuesta:", err)
				break
			}
		
			// Encabezado de tabla
			fmt.Printf("| %-3s | %-10s | %-15s | %-12s | %-6s | %-20s |\n", "#", "ID carrera", "País", "Fecha", "Year", "Circuito")
			fmt.Println(strings.Repeat("-", 80))
		
			for i, c := range carreras {
				// Formatear fecha
				fecha, _ := time.Parse(time.RFC3339, c.DateStart)
				fmt.Printf("| %-3d | %-10d | %-15s | %-12s | %-6d | %-20s |\n",
					i+1, c.SessionKey, c.CountryName, fecha.Format("02-01-2006"), c.Year, c.CircuitShortName)
			}
			fmt.Println(strings.Repeat("-", 80))
		case 4:
			fmt.Println("▶️ [4] Ver detalle de carrera\n")
			fmt.Print("Ingrese el número de la carrera: ")
			var raceID string
			fmt.Scanln(&raceID)
			url := fmt.Sprintf("http://localhost:8080/api/carrera/detalle/%s", raceID)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println(" Error al hacer la solicitud:", err)
				break
			}
			defer resp.Body.Close()
		
			body, _ := io.ReadAll(resp.Body)
		
			if resp.StatusCode != 200 {
				fmt.Println(" Error en la respuesta del servidor:", string(body))
				break
			}
		
			var detalle struct {
				RaceID           string `json:"race_id"`
				CountryName      string `json:"country_name"`
				DateStart        string `json:"date_start"`
				Year             int    `json:"year"`
				CircuitShortName string `json:"circuit_short_name"`
				Results []struct {
					Position interface{} `json:"position"` // <- aquí el cambio
					Driver   string      `json:"driver"`
					Team     string      `json:"team"`
					Country  string      `json:"country"`
				} `json:"results"`
				FastestLap struct {
					Driver     string  `json:"driver"`
					TotalTime  float64 `json:"total_time"`
					Sector1    float64 `json:"sector_1"`
					Sector2    float64 `json:"sector_2"`
					Sector3    float64 `json:"sector_3"`
				} `json:"fastest_lap"`
				MaxSpeed struct {
					Driver   string  `json:"driver"`
					SpeedKMH float64 `json:"speed_kmh"`
				} `json:"max_speed"`
			}
		
			if err := json.Unmarshal(body, &detalle); err != nil {
				fmt.Println("❌ Error al decodificar JSON:", err)
				break
			}
		
			t, _ := time.Parse(time.RFC3339, detalle.DateStart)
			fechaFormateada := t.Format("02-01-2006")
		
			// Header
			fmt.Printf("\n Detalle de carrera: %s (%s)\nFecha: %s | Circuito: %s | Año: %d\n\n",
				detalle.RaceID, detalle.CountryName, fechaFormateada, detalle.CircuitShortName, detalle.Year)
		
			fmt.Println("| Resultados                                                    |")
			fmt.Println("|--------------------------------------------------------------|")
			fmt.Println("| Posicion | Piloto              | Equipo          | Pais      |")
			fmt.Println("|--------------------------------------------------------------|")
			for _, r := range detalle.Results {
				fmt.Printf("| %-8s | %-18s | %-14s | %-9s |\n",fmt.Sprintf("%v", r.Position), r.Driver, r.Team, r.Country)
			}
			fmt.Println("|--------------------------------------------------------------|")
		
			// Vuelta más rápida
			fmt.Println("\n| Vuelta más rápida                                            |")
			fmt.Println("|--------------------------------------------------------------|")
			fmt.Println("| Piloto          | Tiempo Total | Sector 1 | Sector 2 | Sector 3 |")
			fmt.Println("|--------------------------------------------------------------|")
			fmt.Printf("| %-15s | %11s | %8.3f | %8.3f | %8.3f |\n",
				detalle.FastestLap.Driver,
				fmt.Sprintf("%.3f", detalle.FastestLap.TotalTime),
				detalle.FastestLap.Sector1,
				detalle.FastestLap.Sector2,
				detalle.FastestLap.Sector3,
			)
			fmt.Println("|--------------------------------------------------------------|")
		
			// Velocidad máxima
			fmt.Println("\n| Velocidad máxima alcanzada                                   |")
			fmt.Println("|--------------------------------------------------------------|")
			fmt.Println("| Piloto          | Velocidad (km/h)                           |")
			fmt.Println("|--------------------------------------------------------------|")
			fmt.Printf("| %-15s | %-34.1f |\n", detalle.MaxSpeed.Driver, detalle.MaxSpeed.SpeedKMH)
			fmt.Println("|--------------------------------------------------------------|")
		
		case 5:
			fmt.Println(" [5] Ver resumen de temporada\n")
			fmt.Print("Ingrese la temporada (ej: 2024): ")
			var temporada string
			fmt.Scanln(&temporada)
		
			url := "http://localhost:8080/api/temporada/resumen"
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("❌ Error al hacer la solicitud:", err)
				break
			}
			defer resp.Body.Close()
		
			body, _ := io.ReadAll(resp.Body)
		
			if resp.StatusCode != 200 {
				fmt.Println(" Error en la respuesta del servidor:", string(body))
				break
			}
		
			var resumen struct {
				Season          int `json:"season"`
				Top3Winners     []struct {
					Position int    `json:"position"`
					Driver   string `json:"driver"`
					Team     string `json:"team"`
					Country  string `json:"country"`
					Wins     int    `json:"wins"`
				} `json:"top_3_winners"`
				Top3FastestLaps []struct {
					Position    int    `json:"position"`
					Driver      string `json:"driver"`
					Team        string `json:"team"`
					Country     string `json:"country"`
					FastestLaps int    `json:"fastest_laps"`
				} `json:"top_3_fastest_laps"`
				Top3PolePositions []struct {
					Position int    `json:"position"`
					Driver   string `json:"driver"`
					Team     string `json:"team"`
					Country  string `json:"country"`
					Poles    int    `json:"poles"`
				} `json:"top_3_pole_positions"`
			}
		
			if err := json.Unmarshal(body, &resumen); err != nil {
				fmt.Println("❌ Error al decodificar JSON:", err)
				break
			}
		
			fmt.Printf("\n Top 3 Pilotos con más Victorias - Temporada %d\n", resumen.Season)
			fmt.Println("------------------------------------------------------------")
			fmt.Println("| Posición | Piloto           | Equipo         | País | Victorias |")
			fmt.Println("------------------------------------------------------------")
			for _, p := range resumen.Top3Winners {
				fmt.Printf("| %-8d | %-15s | %-14s | %-4s | %-9d |\n",
					p.Position, p.Driver, p.Team, p.Country, p.Wins)
			}
			fmt.Println("------------------------------------------------------------\n")
		
			fmt.Printf(" Top 3 Pilotos con más Vueltas Rápidas - Temporada %d\n", resumen.Season)
			fmt.Println("------------------------------------------------------------")
			fmt.Println("| Posición | Piloto           | Equipo         | País | Vueltas Rápidas |")
			fmt.Println("------------------------------------------------------------")
			for _, p := range resumen.Top3FastestLaps {
				fmt.Printf("| %-8d | %-15s | %-14s | %-4s | %-16d |\n",
					p.Position, p.Driver, p.Team, p.Country, p.FastestLaps)
			}
			fmt.Println("------------------------------------------------------------\n")
		
			fmt.Printf(" Top 3 Pilotos con más Pole Positions - Temporada %d\n", resumen.Season)
			fmt.Println("------------------------------------------------------------")
			fmt.Println("| Posición | Piloto           | Equipo         | País | Poles |")
			fmt.Println("------------------------------------------------------------")
			for _, p := range resumen.Top3PolePositions {
				fmt.Printf("| %-8d | %-15s | %-14s | %-4s | %-5d |\n",
					p.Position, p.Driver, p.Team, p.Country, p.Poles)
			}
			fmt.Println("------------------------------------------------------------\n")
		case 6:
			fmt.Println("👋 Saliendo del programa...")
			return
		default:
			fmt.Println("❌ Opción no reconocida. Intenta nuevamente.\n")
		}
	}
}
