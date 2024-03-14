package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Receipt represents the structure of a receipt
type Receipt struct {
	ID           string        `json:"id,omitempty"`
	Retailer     string        `json:"retailer,omitempty"`
	PurchaseDate string        `json:"purchaseDate,omitempty"`
	PurchaseTime string        `json:"purchaseTime,omitempty"`
	Items        []ReceiptItem `json:"items,omitempty"`
	Total        string        `json:"total,omitempty"`
}

// ReceiptItem represents an item in the receipt
type ReceiptItem struct {
	ShortDescription string `json:"shortDescription,omitempty"`
	Price            string `json:"price,omitempty"`
}

var receipts map[string]Receipt

// ProcessReceiptsEndpoint handles the processing of receipts
func ProcessReceiptsEndpoint(w http.ResponseWriter, req *http.Request) {
	var receipt Receipt
	err := json.NewDecoder(req.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "Failed to decode receipt", http.StatusBadRequest)
		return
	}

	// Generate a unique ID for the receipt
	receipt.ID = uuid.New().String()

	// Store the receipt in memory
	receipts[receipt.ID] = receipt

	// Render a page displaying the ID
	tmpl := template.Must(template.New("receiptID").Parse(`<html><body><h1>Receipt processed successfully!</h1><p>ID: {{ .ID }}</p></body></html>`))
	if err := tmpl.Execute(w, receipt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetPointsEndpoint calculates and returns the points awarded for a receipt
func GetPointsEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	receiptID := params["id"]

	// Retrieve the receipt by ID
	receipt, exists := receipts[receiptID]
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	// Calculate points based on rules
	points := calculatePoints(receipt)

	// Return the points awarded
	json.NewEncoder(w).Encode(map[string]int{"points": points})
}

// calculatePoints calculates the points awarded for a receipt based on the defined rules
func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: One point for every alphanumeric character in the retailer name
	points += len(receipt.Retailer)

	// Rule 2: 50 points if the total is a round dollar amount with no cents
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if total == float64(int(total)) {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25
	totalCents := total * 100
	if int(totalCents)%25 == 0 {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt
	points += len(receipt.Items) / 2 * 5

	// Rule 5: If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer.
	for _, item := range receipt.Items {
		trimmedLength := len(item.ShortDescription)
		if trimmedLength%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int(price * 0.2)
		}
	}

	// Rule 6: 6 points if the day in the purchase date is odd
	purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
	if purchaseDate.Day()%2 != 0 {
		points += 6
	}

	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
	purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
	if purchaseTime.After(time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)) && purchaseTime.Before(time.Date(0, 1, 1, 16, 0, 0, 0, time.UTC)) {
		points += 10
	}

	return points
}

// HomePageHandler serves the home page with a form for JSON input
func HomePageHandler(w http.ResponseWriter, req *http.Request) {
	// Serve an HTML page with a form for JSON input
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Receipt Processing</title>
	</head>
	<body>
		<h1>Receipt Processing</h1>
		<form id="jsonForm" method="post">
			<label for="jsonData">JSON Data:</label>
			<textarea id="jsonData" name="jsonData" rows="10" cols="50" required></textarea><br><br>
			<input type="submit" value="Submit">
		</form>

		<script>
			// JavaScript code to handle form submission
			document.getElementById("jsonForm").addEventListener("submit", function(event) {
				event.preventDefault(); // Prevent the default form submission

				// Get JSON data from the textarea
				var jsonData = document.getElementById("jsonData").value;

				// Send JSON data using fetch API
				fetch('/receipts/process', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: jsonData
				})
				.then(response => response.text())
				.then(data => {
					// Display the ID
					document.body.innerHTML = data;
				})
				.catch(error => {
					console.error('Error:', error);
					alert("Failed to process the receipt. Please try again.");
				});
			});
		</script>
	</body>
	</html>
	`
	fmt.Fprint(w, html)
}

func main() {
	receipts = make(map[string]Receipt)

	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/", HomePageHandler).Methods("GET") // New route for the home page
	router.HandleFunc("/receipts/process", ProcessReceiptsEndpoint).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", GetPointsEndpoint).Methods("GET")

	fmt.Println("Server is running at port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
