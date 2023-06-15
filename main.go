package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"errors"

	"github.com/gin-gonic/gin"
)

// Estructura para representar un valor de referencia
type ReferenceValue struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Reference   float64 `json:"reference"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
}

// Ruta del archivo JSON donde se guardarán los valores de referencia
const dataFile = "reference_values.json"

// Slice de valores de referencia inicial
var referenceValues = []ReferenceValue{}

// Función para cargar los valores de referencia desde el archivo JSON
func loadReferenceValues() error {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &referenceValues)
	if err != nil {
		return err
	}

	return nil
}

// Función para guardar los valores de referencia en el archivo JSON
func saveReferenceValues() error {
	data, err := json.Marshal(referenceValues)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dataFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Handler para obtener todos los valores de referencia
func getReferenceValues(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, referenceValues)
}

// Handler para obtener un valor de referencia por su ID
func getReferenceValueByID(c *gin.Context) {
	id := c.Param("id")
	value, err := findReferenceValueByID(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Value not found."})
		return
	}

	c.IndentedJSON(http.StatusOK, value)
}

// Handler para crear un nuevo valor de referencia
func createReferenceValue(c *gin.Context) {
	var newValue ReferenceValue

	if err := c.BindJSON(&newValue); err != nil {
		return
	}

	referenceValues = append(referenceValues, newValue)

	if err := saveReferenceValues(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to save reference values."})
		return
	}

	c.IndentedJSON(http.StatusCreated, newValue)
}

// Handler para actualizar un valor de referencia existente
func updateReferenceValue(c *gin.Context) {
	id := c.Param("id")
	var updatedValue ReferenceValue

	if err := c.BindJSON(&updatedValue); err != nil {
		return
	}

	for i := range referenceValues {
		if referenceValues[i].ID == id {
			referenceValues[i].Name = updatedValue.Name
			referenceValues[i].Reference = updatedValue.Reference
			referenceValues[i].Description = updatedValue.Description
			referenceValues[i].ImageURL = updatedValue.ImageURL

			if err := saveReferenceValues(); err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to save reference values."})
				return
			}

			c.IndentedJSON(http.StatusOK, referenceValues[i])
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Value not found."})
}

// Función auxiliar para encontrar un valor de referencia por su ID
func findReferenceValueByID(id string) (*ReferenceValue, error) {
	for i := range referenceValues {
		if referenceValues[i].ID == id {
			return &referenceValues[i], nil
		}
	}

	return nil, errors.New("value not found")
}

// Estructura para almacenar los tokens válidos
type TokenStore struct {
	Tokens map[string]bool
}

// Función de middleware para verificar el token en el encabezado de la solicitud
func (ts *TokenStore) AuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")

	if !ts.isValidToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		c.Abort()
		return
	}

	c.Next()
}

// Función para verificar si un token es válido
func (ts *TokenStore) isValidToken(token string) bool {
	_, ok := ts.Tokens[token]
	return ok
}

func main() {
	if err := loadReferenceValues(); err != nil {
		panic(err)
	}

	router := gin.Default()

	// Crea una instancia del almacenamiento de tokens
	tokenStore := &TokenStore{
		Tokens: make(map[string]bool),
	}

	// Agrega tus tokens válidos aquí
	tokenStore.Tokens["25111769"] = true

	// Agrega el middleware de autenticación
	router.Use(tokenStore.AuthMiddleware)

	router.GET("/values", getReferenceValues)
	router.GET("/values/:id", getReferenceValueByID)
	router.POST("/values", createReferenceValue)
	router.PUT("/values/:id", updateReferenceValue)
	router.Run("localhost:8081")
}
