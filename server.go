package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func fetchDataFromAPI(url string) ([]map[string]interface{}, error) {
	// Realizar la solicitud HTTP GET
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al realizar la solicitud: %v", err)
	}
	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer la respuesta: %v", err)
	}

	// Parsear el JSON a una estructura de Go
	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error al parsear el JSON: %v", err)
	}

	return data, nil
}
func nullFloatToFloat(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}

func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
	// Conectar a la base de datos (se crea si no existe)
	db, err := sql.Open("sqlite3", "./proxy.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Crear tabla Driver
	createDriverTable := `
	CREATE TABLE IF NOT EXISTS Driver (
		driver_number INTEGER PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		name_acronym TEXT NOT NULL,
		team_name TEXT NOT NULL,
		country_code TEXT NOT NULL
	);`

	// Crear tabla Session
	createSessionTable := `
	CREATE TABLE IF NOT EXISTS Session (
		session_key INTEGER PRIMARY KEY,
		session_name TEXT NOT NULL,
		session_type TEXT NOT NULL,
		location TEXT NOT NULL,
		country_name TEXT NOT NULL,
		year INTEGER NOT NULL,
		circuit_short_name TEXT NOT NULL,
		date_start TEXT NOT NULL
	);`

	// Crear tabla Position
	createPositionTable := `
	CREATE TABLE IF NOT EXISTS Position (
		driver_number INTEGER NOT NULL,
		session_key INTEGER NOT NULL,
		position INTEGER NOT NULL,
		date TEXT NOT NULL,
		PRIMARY KEY (driver_number, session_key)
		FOREIGN KEY (driver_number) REFERENCES Driver(driver_number),
		FOREIGN KEY (session_key) REFERENCES Session(session_key)
	);`

	// Crear tabla Laps
	createLapsTable := `
	CREATE TABLE IF NOT EXISTS Laps (
		driver_number INTEGER NOT NULL,
		session_key INTEGER NOT NULL,
		lap_number INTEGER NOT NULL,
		lap_duration REAL NOT NULL,
		duration_sector_1 REAL NOT NULL,
		duration_sector_2 REAL NOT NULL,
		duration_sector_3 REAL NOT NULL,
		st_speed REAL NOT NULL,
		date_start TEXT NOT NULL,
		PRIMARY KEY (driver_number, session_key, lap_number),
		FOREIGN KEY (driver_number) REFERENCES Driver(driver_number),
		FOREIGN KEY (session_key) REFERENCES Session(session_key)
	);`

	// Ejecutar las sentencias SQL
	tables := map[string]string{
		"Driver":   createDriverTable,
		"Session":  createSessionTable,
		"Position": createPositionTable,
		"Laps":     createLapsTable,
	}

	for tableName, query := range tables {
		_, err = db.Exec(query)
		if err != nil {
			log.Fatalf("Error al crear la tabla %s: %v", tableName, err)
		}
		fmt.Printf("Tabla %s creada exitosamente\n", tableName)
	}

	fmt.Println("Todas las tablas fueron creadas correctamente")

	//----------------------------------------------------------------------
	// 1. Rellenar la tabla de pilotos:

	// URL de la API
	url := "https://api.openf1.org/v1/drivers?session_key=9574"

	insertDriver := `
	INSERT OR IGNORE INTO Driver (driver_number, first_name, last_name, name_acronym, team_name, country_code)
	VALUES (?, ?, ?, ?, ?, ?)`

	// Realizar la consulta a la API
	data, err := fetchDataFromAPI(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	numbers_drivers_to_extract := []int{1, 2, 3, 4, 10, 11, 14, 16, 18, 20, 22, 23, 24, 27, 31, 44, 55, 63, 77, 81}

	// Extraer drivers pedidos
	fmt.Println("Datos obtenidos de la API:")
	for _, driver := range data {
		driverNumber := int(driver["driver_number"].(float64)) // Convertir de float64 a int
		if contains(numbers_drivers_to_extract, driverNumber) {
			fmt.Printf("- %s (%s) %d\n", driver["first_name"], driver["team_name"], driverNumber)
			_, err = db.Exec(insertDriver, driverNumber, driver["first_name"], driver["last_name"], driver["name_acronym"], driver["team_name"], driver["country_code"])
			if err != nil {
				log.Fatal("Error insertando driver:", err)
			}
			fmt.Println("Driver insertado correctamente")
		}
	}

	// URL de la API
	url = "https://api.openf1.org/v1/drivers?session_key=9636"

	// Realizar la consulta a la API
	data, err = fetchDataFromAPI(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	numbers_drivers_to_extract = []int{30, 50, 43}

	// Extraer drivers pedidos
	fmt.Println("Datos obtenidos de la API:")
	for _, driver := range data {
		driverNumber := int(driver["driver_number"].(float64)) // Convertir de float64 a int
		if contains(numbers_drivers_to_extract, driverNumber) {
			fmt.Printf("- %s (%s) %d\n", driver["first_name"], driver["team_name"], driverNumber)
			_, err = db.Exec(insertDriver, driverNumber, driver["first_name"], driver["last_name"], driver["name_acronym"], driver["team_name"], driver["country_code"])
			if err != nil {
				log.Fatal("Error insertando driver:", err)
			}
			fmt.Println("Driver insertado correctamente")
		}
	}
	//----------------------------------------------------------------------
	// 2. Rellenar tabla de carreras:
	// URL de la API
	url = "https://api.openf1.org/v1/sessions?session_name=Race&year=2024"

	insertSession := `
	INSERT OR IGNORE INTO session (session_key, session_name, session_type, location, country_name, year, circuit_short_name, date_start)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	// Realizar la consulta a la API
	data, err = fetchDataFromAPI(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Extraer session pedidos
	for _, session := range data {
		sessionKey := int(session["session_key"].(float64)) // Convertir de float64 a int
		year := int(session["year"].(float64))              // Convertir de float64 a int
		fmt.Printf("- %d (%s) %s %s %s %d %s %s\n", sessionKey, session["session_name"], session["session_type"], session["location"], session["country_name"], year, session["circuit_short_name"], session["date_start"])
		_, err = db.Exec(insertSession, sessionKey, session["session_name"], session["session_type"], session["location"], session["country_name"], year, session["circuit_short_name"], session["date_start"])
		if err != nil {
			log.Fatal("Error insertando session:", err)
		}
		fmt.Println("Session insertado correctamente")
	}
	//----------------------------------------------------------------------
	// 3. Rellenar tabla de posiciones

	// Configurar SQLite para mejor manejo de concurrencia
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("Error configurando WAL mode: %v", err)
	}
	_, err = db.Exec("PRAGMA busy_timeout=10000;")
	if err != nil {
		log.Printf("Error configurando busy timeout: %v", err)
	}

	// Obtener todas las session_keys de la base de datos
	rows, err := db.Query("SELECT session_key FROM Session")
	if err != nil {
		log.Fatal("Error consultando session_keys:", err)
	}

	var sessionKeys []int
	for rows.Next() {
		var sessionKey int
		if err := rows.Scan(&sessionKey); err != nil {
			log.Fatal("Error escaneando session_key:", err)
		}
		sessionKeys = append(sessionKeys, sessionKey)
	}
	rows.Close()

	// Función mejorada para reintentos en caso de fallo
	retryOperation := func(operation func() error, maxRetries int) error {
		var err error
		for i := 0; i < maxRetries; i++ {
			err = operation()
			if err == nil {
				return nil
			}

			if strings.Contains(err.Error(), "database is locked") {
				waitTime := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
				log.Printf("Intento %d fallido: %v. Esperando %v antes de reintentar...", i+1, err, waitTime)
				time.Sleep(waitTime)
				continue
			}
			return err
		}
		return fmt.Errorf("después de %d reintentos: %v", maxRetries, err)
	}

	// Procesar cada session
	for _, sessionKey := range sessionKeys {
		fmt.Printf("\nProcesando posiciones para session_key=%d\n", sessionKey)

		// Consultar la API para obtener todas las posiciones
		positionsURL := fmt.Sprintf("https://api.openf1.org/v1/position?session_key=%d", sessionKey)
		positionsData, err := fetchDataFromAPI(positionsURL)
		if err != nil {
			log.Printf("Error obteniendo posiciones para session_key=%d: %v", sessionKey, err)
			continue
		}

		fmt.Printf("Obtenidos %d registros de posición\n", len(positionsData))

		// Procesar en lotes más pequeños
		batchSize := 100
		totalProcessed := 0

		for i := 0; i < len(positionsData); i += batchSize {
			end := i + batchSize
			if end > len(positionsData) {
				end = len(positionsData)
			}
			batch := positionsData[i:end]

			err := retryOperation(func() error {
				tx, err := db.Begin()
				if err != nil {
					return fmt.Errorf("error al iniciar transacción: %v", err)
				}

				stmt, err := tx.Prepare(`
                INSERT OR IGNORE INTO Position 
                (driver_number, session_key, position, date) 
                VALUES (?, ?, ?, ?)`)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error preparando statement: %v", err)
				}
				defer stmt.Close()

				for _, pos := range batch {
					driverNumber := int(pos["driver_number"].(float64))

					if driverNumber == 61 {
						continue
					}

					position := int(pos["position"].(float64))
					date := pos["date"].(string)

					_, err = stmt.Exec(driverNumber, sessionKey, position, date)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("error insertando posición: %v", err)
					}
				}

				if err := tx.Commit(); err != nil {
					return fmt.Errorf("error en commit: %v", err)
				}
				return nil
			}, 5) // 5 reintentos máximo

			if err != nil {
				log.Printf("Error persistente con el lote %d-%d: %v", i, end, err)
				continue
			}

			totalProcessed += len(batch)
			fmt.Printf("Procesados %d/%d registros (%.1f%%)\n",
				totalProcessed, len(positionsData),
				float64(totalProcessed)/float64(len(positionsData))*100)
		}

		fmt.Printf("Finalizado session_key=%d. Total procesados: %d/%d\n",
			sessionKey, totalProcessed, len(positionsData))
	}

	//Volviendo a configuración inicial
	_, err = db.Exec("PRAGMA journal_mode=DELETE;")
	if err != nil {
		log.Printf("Error configurando modo DELETE: %v", err)
	}

	_, err = db.Exec("PRAGMA busy_timeout=0;")
	if err != nil {
		log.Printf("Error configurando busy timeout a 0: %v", err)
	}

	//----------------------------------------------------------------------
	// 4. Rellenar tabla de vueltas:

	// Mantener la configuración WAL para mejor rendimiento con la tabla de vueltas
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("Error configurando WAL mode para laps: %v", err)
	}
	_, err = db.Exec("PRAGMA busy_timeout=10000;")
	if err != nil {
		log.Printf("Error configurando busy timeout para laps: %v", err)
	}

	fmt.Println("\nComenzando procesamiento de vueltas...")

	for _, sessionKey := range sessionKeys {
		fmt.Printf("\nProcesando vueltas para session_key=%d\n", sessionKey)

		// Consultar la API para obtener todas las laps de esta session
		lapsURL := fmt.Sprintf("https://api.openf1.org/v1/laps?session_key=%d", sessionKey)
		lapsData, err := fetchDataFromAPI(lapsURL)
		if err != nil {
			log.Printf("Error obteniendo vueltas para session_key=%d: %v", sessionKey, err)
			continue
		}

		fmt.Printf("Obtenidos %d registros de vueltas\n", len(lapsData))

		// Procesar en lotes más pequeños
		batchSize := 100
		totalProcessed := 0

		for i := 0; i < len(lapsData); i += batchSize {
			end := i + batchSize
			if end > len(lapsData) {
				end = len(lapsData)
			}
			batch := lapsData[i:end]

			err := retryOperation(func() error {
				tx, err := db.Begin()
				if err != nil {
					return fmt.Errorf("error al iniciar transacción para laps: %v", err)
				}

				stmt, err := tx.Prepare(`
            INSERT OR IGNORE INTO Laps 
            (driver_number, session_key, lap_number, lap_duration, 
             duration_sector_1, duration_sector_2, duration_sector_3, 
             st_speed, date_start) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error preparando statement para laps: %v", err)
				}
				defer stmt.Close()

				for _, lap := range batch {
					// Verificar que los campos obligatorios existan
					if lap["driver_number"] == nil || lap["lap_number"] == nil {
						continue
					}

					driverNumber := int(lap["driver_number"].(float64))

					if driverNumber == 61 {
						continue
					}

					lapNumber := int(lap["lap_number"].(float64))

					// Manejar campos que podrían ser nulos
					var lapDuration float64
					var durationSector1 float64
					var durationSector2 float64
					var durationSector3 float64
					var stSpeed float64
					var dateStart string

					lapDuration = 0
					if lap["lap_duration"] != nil {
						lapDuration = lap["lap_duration"].(float64)
					} else {
						if lap["duration_sector_1"] != nil {
							lapDuration += lap["duration_sector_1"].(float64)
						}
						if lap["duration_sector_2"] != nil {
							lapDuration += lap["duration_sector_2"].(float64)
						}
						if lap["duration_sector_3"] != nil {
							lapDuration += lap["duration_sector_3"].(float64)
						}
					}
					if lap["duration_sector_1"] != nil {
						durationSector1 = lap["duration_sector_1"].(float64)
					}
					if lap["duration_sector_2"] != nil {
						durationSector2 = lap["duration_sector_2"].(float64)
					}
					if lap["duration_sector_3"] != nil {
						durationSector3 = lap["duration_sector_3"].(float64)
					}
					if lap["st_speed"] != nil {
						stSpeed = lap["st_speed"].(float64)
					}

					// Manejar el caso específico de date_start
					if lap["date_start"] != nil {
						dateStart = lap["date_start"].(string)
					} else {
						// Usar una fecha por defecto o la fecha actual si date_start es nil
						// Revisar que hacer
						dateStart = time.Now().UTC().Format(time.RFC3339)
						//////////////////
					}

					_, err = stmt.Exec(
						driverNumber, sessionKey, lapNumber,
						lapDuration, durationSector1, durationSector2,
						durationSector3, stSpeed, dateStart)

					if err != nil {
						tx.Rollback()
						return fmt.Errorf("error insertando vuelta: %v", err)
					}
				}

				if err := tx.Commit(); err != nil {
					return fmt.Errorf("error en commit para laps: %v", err)
				}
				return nil
			}, 5) // 5 reintentos máximo

			if err != nil {
				log.Printf("Error persistente con el lote de vueltas %d-%d: %v", i, end, err)
				continue
			}

			totalProcessed += len(batch)
			fmt.Printf("Procesados %d/%d registros de vueltas (%.1f%%)\n",
				totalProcessed, len(lapsData),
				float64(totalProcessed)/float64(len(lapsData))*100)
		}

		fmt.Printf("Finalizado procesamiento de vueltas para session_key=%d. Total procesados: %d/%d\n",
			sessionKey, totalProcessed, len(lapsData))
	}

	// Volviendo a configuración inicial
	_, err = db.Exec("PRAGMA journal_mode=DELETE;")
	if err != nil {
		log.Printf("Error configurando modo DELETE después de procesar vueltas: %v", err)
	}

	_, err = db.Exec("PRAGMA busy_timeout=0;")
	if err != nil {
		log.Printf("Error configurando busy timeout a 0 después de procesar vueltas: %v", err)
	}

	fmt.Println("Procesamiento de vueltas completado")

	//----------------------------------------------------------------------
	// Servidor

	r := gin.Default()

	r.GET("/api/corredor", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT first_name, last_name, driver_number, team_name, country_code
			FROM Driver
			ORDER BY driver_number ASC
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error al consultar los corredores"})
			return
		}
		defer rows.Close()
	
		var corredores []gin.H
	
		for rows.Next() {
			var firstName, lastName, teamName, countryCode string
			var driverNumber int
	
			if err := rows.Scan(&firstName, &lastName, &driverNumber, &teamName, &countryCode); err != nil {
				c.JSON(500, gin.H{"error": "Error al leer resultados"})
				return
			}
	
			corredores = append(corredores, gin.H{
				"first_name":    firstName,
				"last_name":     lastName,
				"driver_number": driverNumber,
				"team_name":     teamName,
				"country_code":  countryCode,
			})
		}
	
		c.JSON(200, corredores)
	})



	r.GET("/api/corredor/detalle/:id", func(c *gin.Context) {
		driverID := c.Param("id")
	
		// 1. Obtener carreras ganadas y top 3
		var wins, top3 int
		err := db.QueryRow(`
			SELECT 
				COUNT(DISTINCT CASE WHEN position = 1 THEN session_key END),
				COUNT(DISTINCT CASE WHEN position <= 3 THEN session_key END)
			FROM Position
			WHERE driver_number = ?
		`, driverID).Scan(&wins, &top3)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error obteniendo victorias/top3"})
			return
		}
	
		// 2. Velocidad máxima
		var maxSpeed float64
		err = db.QueryRow(`
			SELECT MAX(st_speed)
			FROM Laps
			WHERE driver_number = ?
		`, driverID).Scan(&maxSpeed)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error obteniendo velocidad máxima"})
			return
		}
	
		// 3. Resultados por carrera (una por session_key)
		rows, err := db.Query(`
		SELECT 
			s.session_key,
			s.circuit_short_name,
			s.country_name,
			MIN(p.position) AS position,
			(
				SELECT MIN(lap_duration)
				FROM Laps
				WHERE driver_number = p.driver_number
				AND session_key = p.session_key
				AND lap_duration > 0
			) AS best_lap_duration,
			(
				SELECT MAX(st_speed)
				FROM Laps
				WHERE driver_number = p.driver_number
				AND session_key = p.session_key
			) AS max_speed,
			CASE
				WHEN (
					SELECT MIN(lap_duration)
					FROM Laps
					WHERE session_key = p.session_key
					AND lap_duration > 0
				) = (
					SELECT MIN(lap_duration)
					FROM Laps
					WHERE session_key = p.session_key
					AND driver_number = p.driver_number
					AND lap_duration > 0
				)
				THEN true
				ELSE false
			END AS fastest_lap
		FROM Position p
		JOIN Session s ON s.session_key = p.session_key
		WHERE p.driver_number = ?
		GROUP BY s.session_key
		ORDER BY s.date_start ASC	
		`, driverID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error consultando resultados del piloto"})
			return
		}
		defer rows.Close()
	
		var resultados []gin.H
		for rows.Next() {
			var sessionKey int
			var circuito, pais string
			var position int
			var bestLap, maxVel sql.NullFloat64
			var fastestLap bool
	
			err := rows.Scan(&sessionKey, &circuito, &pais, &position, &bestLap, &maxVel, &fastestLap)
			if err != nil {
				c.JSON(500, gin.H{"error": "Error leyendo datos de carrera"})
				return
			}
	
			resultados = append(resultados, gin.H{
				"session_key":        sessionKey,
				"circuit_short_name": circuito,
				"race":               "GP de " + pais,
				"position":           position,
				"fastest_lap":        fastestLap,
				"max_speed":          nullFloatToFloat(maxVel),
				"best_lap_duration":  nullFloatToFloat(bestLap),
			})
		}
	
		// 4. Estructura final de respuesta
		c.JSON(200, gin.H{
			"driver_id": driverID,
			"performance_summary": gin.H{
				"wins":        wins,
				"top_3_finishes": top3,
				"max_speed":   maxSpeed,
			},
			"race_results": resultados,
		})
	})
	
	

	r.GET("/api/carrera", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT session_key, country_name, date_start, year, circuit_short_name
			FROM Session
			WHERE session_name = 'Race'
			ORDER BY date_start ASC
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error al consultar las carreras"})
			return
		}
		defer rows.Close()
	
		var carreras []gin.H
	
		for rows.Next() {
			var sessionKey int
			var countryName, dateStart, circuitShortName string
			var year int
	
			if err := rows.Scan(&sessionKey, &countryName, &dateStart, &year, &circuitShortName); err != nil {
				c.JSON(500, gin.H{"error": "Error al leer resultados"})
				return
			}
	
			carreras = append(carreras, gin.H{
				"session_key":        sessionKey,
				"country_name":       countryName,
				"date_start":         dateStart,
				"year":               year,
				"circuit_short_name": circuitShortName,
			})
		}
	
		c.JSON(200, carreras)
	})
	

	
	r.GET("/api/carrera/detalle/:id", func(c *gin.Context) {
		sessionID := c.Param("id")
	
		// 1. Info general
		var country, date, circuit string
		var year int
		err := db.QueryRow(`
			SELECT country_name, date_start, year, circuit_short_name
			FROM Session WHERE session_key = ?
		`, sessionID).Scan(&country, &date, &year, &circuit)
		if err != nil {
			c.JSON(500, gin.H{"error": "Carrera no encontrada"})
			return
		}
	
		// 2. Podio
		podioRows, _ := db.Query(`
			SELECT p.position, d.first_name || ' ' || d.last_name, d.team_name, d.country_code
			FROM Position p
			JOIN Driver d ON d.driver_number = p.driver_number
			WHERE p.session_key = ?
			ORDER BY p.position ASC
			LIMIT 3
		`, sessionID)
		var podio []gin.H
		for podioRows.Next() {
			var pos int
			var name, team, country string
			podioRows.Scan(&pos, &name, &team, &country)
			podio = append(podio, gin.H{
				"position": pos,
				"driver":   name,
				"team":     team,
				"country":  country,
			})
		}
	
		// 3. Último lugar
		var lastPos int
		var lastDriver, lastTeam, lastCountry string
		db.QueryRow(`
			SELECT p.position, d.first_name || ' ' || d.last_name, d.team_name, d.country_code
			FROM Position p
			JOIN Driver d ON d.driver_number = p.driver_number
			WHERE p.session_key = ?
			ORDER BY p.position DESC
			LIMIT 1
		`, sessionID).Scan(&lastPos, &lastDriver, &lastTeam, &lastCountry)
	
		// 4. Vuelta rápida
		var fastDriver string
		var lapTime, sec1, sec2, sec3 float64
		db.QueryRow(`
			SELECT d.first_name || ' ' || d.last_name, lap_duration, duration_sector_1, duration_sector_2, duration_sector_3
			FROM Laps l
			JOIN Driver d ON d.driver_number = l.driver_number
			WHERE l.session_key = ? AND lap_duration > 0
			ORDER BY lap_duration ASC
			LIMIT 1
		`, sessionID).Scan(&fastDriver, &lapTime, &sec1, &sec2, &sec3)
	
		// 5. Velocidad máxima
		var maxDriver string
		var maxSpeed float64
		db.QueryRow(`
			SELECT d.first_name || ' ' || d.last_name, MAX(l.st_speed)
			FROM Laps l
			JOIN Driver d ON d.driver_number = l.driver_number
			WHERE l.session_key = ?
		`, sessionID).Scan(&maxDriver, &maxSpeed)
	
		// 🧾 Estructura de respuesta
		c.JSON(200, gin.H{
			"race_id":           sessionID,
			"country_name":      country,
			"date_start":        date,
			"year":              year,
			"circuit_short_name": circuit,
			"results":           append(podio, gin.H{
				"position": "Último",
				"driver":   lastDriver,
				"team":     lastTeam,
				"country":  lastCountry,
			}),
			"fastest_lap": gin.H{
				"driver":     fastDriver,
				"total_time": lapTime,
				"sector_1":   sec1,
				"sector_2":   sec2,
				"sector_3":   sec3,
			},
			"max_speed": gin.H{
				"driver":    maxDriver,
				"speed_kmh": maxSpeed,
			},
		})
	})

	



	r.GET("/api/temporada/resumen", func(c *gin.Context) {
		// 1. Top 3 ganadores
		winnersRows, err := db.Query(`
			SELECT 
				d.first_name || ' ' || d.last_name AS driver,
				d.team_name,
				d.country_code,
				COUNT(*) AS wins
			FROM Position p
			JOIN Session s ON p.session_key = s.session_key
			JOIN Driver d ON p.driver_number = d.driver_number
			WHERE s.year = 2024 AND s.session_name = 'Race' AND p.position = 1
			GROUP BY d.driver_number
			ORDER BY wins DESC
			LIMIT 3;
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error al obtener ganadores"})
			return
		}
		var topWinners []gin.H
		i := 1
		for winnersRows.Next() {
			var driver, team, country string
			var wins int
			winnersRows.Scan(&driver, &team, &country, &wins)
			topWinners = append(topWinners, gin.H{
				"position": i,
				"driver":   driver,
				"team":     team,
				"country":  country,
				"wins":     wins,
			})
			i++
		}
	
		// 2. Top 3 vueltas rápidas
		fastestRows, err := db.Query(`
			WITH fastest_laps AS (
				SELECT session_key, driver_number, MIN(lap_duration) AS min_time
				FROM Laps
				WHERE lap_duration > 0
				  AND duration_sector_1 > 0 AND duration_sector_2 > 0 AND duration_sector_3 > 0
				GROUP BY session_key
			)
			SELECT 
				d.first_name || ' ' || d.last_name AS driver,
				d.team_name,
				d.country_code,
				COUNT(*) AS fastest_laps
			FROM fastest_laps fl
			JOIN Session s ON fl.session_key = s.session_key
			JOIN Driver d ON fl.driver_number = d.driver_number
			WHERE s.year = 2024 AND s.session_name = 'Race'
			GROUP BY d.driver_number
			ORDER BY fastest_laps DESC
			LIMIT 3;
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error al obtener vueltas rápidas"})
			return
		}
		var topFastest []gin.H
		i = 1
		for fastestRows.Next() {
			var driver, team, country string
			var laps int
			fastestRows.Scan(&driver, &team, &country, &laps)
			topFastest = append(topFastest, gin.H{
				"position":     i,
				"driver":       driver,
				"team":         team,
				"country":      country,
				"fastest_laps": laps,
			})
			i++
		}
	
		// 3. Top 3 en podios (corredores con más posiciones <= 3 en cualquier carrera)
		podiumRows, err := db.Query(`
		SELECT 
			d.first_name || ' ' || d.last_name AS driver,
			d.team_name,
			d.country_code,
			COUNT(DISTINCT p.session_key) AS podiums
		FROM Position p
		JOIN Session s ON p.session_key = s.session_key
		JOIN Driver d ON p.driver_number = d.driver_number
		WHERE s.year = 2024 AND p.position <= 3
		GROUP BY d.driver_number
		ORDER BY podiums DESC
		LIMIT 3;
		`)
		if err != nil {
		c.JSON(500, gin.H{"error": "Error al obtener top 3 en podios"})
		return
		}
		var topPodiums []gin.H
		i = 1
		for podiumRows.Next() {
		var driver, team, country string
		var count int
		err := podiumRows.Scan(&driver, &team, &country, &count)
		if err != nil {
			log.Printf("Error escaneando podio: %v", err)
			continue
		}
		topPodiums = append(topPodiums, gin.H{
			"position": i,
			"driver":   driver,
			"team":     team,
			"country":  country,
			"podiums":  count,
		})
		i++
}
		// 4. Respuesta final
		c.JSON(200, gin.H{
			"season":               2024,
			"top_3_winners":        topWinners,
			"top_3_fastest_laps":   topFastest,
			"top_3_pole_positions": topPodiums, // Ahora bien definido como top 3 clasificadores
		})
	})
	
	

	r.Run(":8080")
}