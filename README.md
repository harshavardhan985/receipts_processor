# receipts_processor
Technologies: Go, Docker

End Points:
Path: localhost:8080
Method: GET
1) It will land on homepage with form that takes json format of receipt.

Path: localhost:8080/receipts/process
Method: POST
Payload: Receipt JSON
Response: JSON containing an id for the receipt.

Path: localhost:8080/receipts/{id}/points
Method: GET
Response: A JSON object containing the number of points awarded.

