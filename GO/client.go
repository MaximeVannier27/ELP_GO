// client.go

package main

import (
	//"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"

	//"io"
	"log"
	"net"
	"os"
)

func main() {
	// Se connecter au serveur sur le port 8080
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erreur lors de la connexion au serveur:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connexion établie avec le serveur!")

	// Ouvrir le fichier de l'image à envoyer
	file, err := os.Open("CGR.jpg")
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier:", err)
		return
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	//var buffer bytes.Buffer
	err = jpeg.Encode(conn, image, nil)
	if err != nil {
		log.Fatal((err))
	}

	// RECEPTION DU RESULTAT ET RESTRUCTURATION EN JPEG

	image_finale, err := jpeg.Decode(conn)
	if err != nil {
		log.Fatal(err)
	}

	fichier, err := os.Create("FLOU_client.jpeg")
	if err != nil {
		panic(err)
	}
	defer fichier.Close()

	err = jpeg.Encode(fichier, image_finale, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Image traitée sauvegardée avec succès.")
}
