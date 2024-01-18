// server.go

package main

import (
	//"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"

	//"io"
	"bufio"
	"log"
	"net"
	"strconv"
)

type Pixel struct {
	Red   uint32
	Green uint32
	Blue  uint32
	Alpha uint32
	Coord [2]int // {x,y}
}

type Image struct {
	Width   int
	Height  int
	Radius  int
	Matrix  [][]Pixel
	Channel [6]uint32
}

func uint32ToUint8(value uint32) uint8 {
	scaledValue := float64(value) * (255.0 / 0xffff)
	return uint8(scaledValue)
}

func initImage(image image.Image, r_floutage int, ch_res [6]uint32) Image {

	bordures := image.Bounds()

	var rayon int
	var tmp []Pixel
	var channel [6]uint32

	fmt.Print("Rayon de floutage: ")
	fmt.Scanln(&rayon)

	rayon = r_floutage
	channel = ch_res

	retour := Image{bordures.Max.X, bordures.Max.Y, rayon, [][]Pixel{}, channel}
	for y := bordures.Min.Y; y < bordures.Max.Y; y++ {
		tmp = nil
		for x := bordures.Min.X; x < bordures.Max.X; x++ {

			r, g, b, a := image.At(x, y).RGBA()
			p := Pixel{r, g, b, a, [2]int{x, y}}
			tmp = append(tmp, p)
		}
		retour.Matrix = append(retour.Matrix, tmp)
	}
	return retour
}

func franceTravail(jobs chan<- [2]int, height int, num_image int) {
	for i := 0; i < height; i++ {
		jobs <- [2]int{i, num_image}
	}
	return
}

func interimaire(jobs <-chan [2]int) {
	var im_in Image
	var r int
	var red_avg, green_avg, blue_avg, alpha_avg, comp uint32
	var envoi [6]uint32
	for {
		for contenu := range jobs {
			y_im := contenu[0]
			num_im := contenu[1]
			im_in = dict_images[num_im]
			res := im_in.Channel
			r = im_in.Radius
			for x_im := 0; x_im < im_in.Width; x_im++ {
				red_avg, green_avg, blue_avg, alpha_avg, comp = 0, 0, 0, 0, 0
				for y_pix := 0; y_pix < 2*r+1; y_pix++ {
					for x_pix := 0; x_pix < 2*r+1; x_pix++ {
						if y_im+y_pix-r >= 0 && y_im+y_pix-r < im_in.Height && x_im+x_pix-r >= 0 && x_im+x_pix-r < im_in.Width {
							red_avg += (im_in.Matrix[y_pix+y_im-r][x_im+x_pix-r]).Red
							green_avg += (im_in.Matrix[y_pix+y_im-r][x_im+x_pix-r]).Green
							blue_avg += (im_in.Matrix[y_pix+y_im-r][x_im+x_pix-r]).Blue
							alpha_avg += (im_in.Matrix[y_pix+y_im-r][x_im+x_pix-r]).Alpha
							comp++
						}
					}
				}
				envoi[0] = uint32(x_im)
				envoi[1] = uint32(y_im)
				envoi[2] = red_avg / comp
				envoi[3] = green_avg / comp
				envoi[4] = blue_avg / comp
				envoi[5] = alpha_avg / comp
				res <- envoi
			}
		}
	}
}

func handleConnection(conn net.Conn) {
	// Traitement de la connexion ici
	fmt.Println("Nouvelle connexion établie!")

	fmt.Fprintf(conn, "Veuillez entrer le rayon de floutage : ")

	// Lire la réponse du client
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	rayonFloutageStr := scanner.Text()

	// Convertir la réponse en entier
	rayonFloutage, err := strconv.Atoi(rayonFloutageStr)
	if err != nil {
		fmt.Println("Erreur lors de la conversion en entier :", err)
		return
	}

	// Afficher le rayon de floutage
	fmt.Println("Rayon de floutage reçu du client :", rayonFloutage)

	// Fermer la connexion quand c'est terminé
	defer conn.Close()

	// Créer un buffer pour recevoir l'image en bytes
	//var buffer bytes.Buffer
	//io.Copy(&buffer, conn)
	// Retransformer l'image depuis le buffer

	image, _, err := image.Decode(conn)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Creation de l'objet image

	ch_res := make(chan [6]uint32)
	im := initImage(image, rayonFloutage, ch_res)
	c := 0
	var pixel [6]uint32
	im_out := image.NewRGBA(image.Rect(0, 0, im.Width, im.Height))
	i := 0
	for {
		if _, exists := dict_images[i]; exists {
			i++
		} else {
			break
		}
	}
	dict_images[i] = im

	go franceTravail(ch_travail, im.Height, i)

	for c < im.Height*im.Width {
		pixel = <-ch_res
		im_out.Set(int(pixel[0]), int(pixel[1]), color.RGBA{uint32ToUint8(pixel[2]), uint32ToUint8(pixel[3]), uint32ToUint8(pixel[4]), uint32ToUint8(pixel[5])})
		c++
	}
	filename := fmt.Sprintf("image_client%d.png", i)
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, im_out)
	if err != nil {
		panic(err)
	}

	delete(dict_images, i)
	close(ch_res)

}

func main() {
	// Écoute sur le port 8080
	dict_images := make(map[int]int)
	ch_travail := make(chan [2]int)

	for i := 0; i < 16; i++ {
		go interimaire(ch_travail)
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Erreur lors de l'écoute:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Serveur en attente de connexions sur le port 8080...")

	for {
		// Attendre une nouvelle connexion
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erreur lors de l'acceptation de la connexion:", err)
			return
		}

		// Gérer la connexion dans une goroutine pour permettre la gestion de plusieurs connexions simultanées
		go handleConnection(conn)
	}
}
