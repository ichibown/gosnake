package main

import (
	"image"
	_ "image/jpeg"
	"log"
	"math/rand"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/app/debug"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/glsprite"
)

var (
	textures = make(map[int]sprite.SubTex)
	eng      = glsprite.Engine()
	scene    *sprite.Node
	ticker   *time.Ticker

	NODE_SIZE float32

	rowCount int
	colCount int

	head      *SnakeNode
	food      *Node
	direction int
)

const (
	// textures
	FOOD = iota
	BODY
	HEAD

	// directions
	UP
	DOWN
	LEFT
	RIGHT
)

func main() {
	app.Run(app.Callbacks{
		Start: onStart,
		Draw:  onDraw,
		Touch: onTouch,
		Stop:  onStop,
	})
}

func onStart() {
	rand.Seed(time.Now().UTC().UnixNano())
	initTextures()
	initScene()
	initTicker()
	initSnake()
	log.Println("application start")
}

func onStop() {
	ticker.Stop()
	log.Println("application stop")
}

func onTouch(t event.Touch) {
	x := t.Loc.X
	y := t.Loc.Y
	width := geom.Width
	height := geom.Height
	switch {
	case x < width/3 && y > height/3 && y < height*2/3 && direction != RIGHT:
		direction = LEFT
	case y < height/3 && x > width/3 && x < width*2/3 && direction != DOWN:
		direction = UP
	case x > width*2/3 && y > height/3 && y < height*2/3 && direction != LEFT:
		direction = RIGHT
	case y > height*2/3 && x > width/3 && x < width*2/3 && direction != UP:
		direction = DOWN
	}
}

func onDraw() {
	// draw background.
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	// render root node (scene). then it will render whole node tree.
	eng.Render(scene, 0)
	debug.DrawFPS()
}

func tick() {
	snakeMove()
}

func snakeMove() {
	preX := head.Value.X
	preY := head.Value.Y
	x := preX
	y := preY
	switch direction {
	case LEFT:
		x = x - 1
	case UP:
		y = y - 1
	case RIGHT:
		x = x + 1
	case DOWN:
		y = y + 1
	}
	if x < 0 {
		x = colCount
	}
	if x > colCount {
		x = 0
	}
	if y < 0 {
		y = rowCount
	}
	if y > rowCount {
		y = 0
	}
	if checkCanEat(x, y) {
		snakeEat(x, y)
		return
	} else {
		head.Value.SetLocation(x, y)
	}
	pNode := head.Next
	for pNode != nil && pNode.Value != nil {
		tempX := pNode.Value.X
		tempY := pNode.Value.Y
		pNode.Value.SetLocation(preX, preY)
		preX = tempX
		preY = tempY
		pNode = pNode.Next
	}
}

func initSnake() {
	NODE_SIZE = geom.PixelsPerPt
	colCount = int(float32(geom.Width) / NODE_SIZE)
	rowCount = int(float32(geom.Height) / NODE_SIZE)
	direction = LEFT

	food = newNode(FOOD)
	food.SetLocation(rand.Intn(colCount), rand.Intn(rowCount))
	head = newSnakeNode(HEAD)
	head.Value.SetLocation(colCount/2, rowCount/2)
}

func initTicker() {
	ticker = time.NewTicker(time.Second / 2)
	go func() {
		for range ticker.C {
			tick()
		}
	}()
}

func checkCanEat(x, y int) bool {
	return x == food.X && y == food.Y
}

func snakeEat(x, y int) {
	newNode := newSnakeNode(BODY)
	newNode.Value.SetLocation(head.Value.X, head.Value.Y)
	newNode.Next = head.Next
	head.Next = newNode
	head.Value.SetLocation(x, y)
	food.SetLocation(rand.Intn(colCount), rand.Intn(rowCount))
}

func initScene() {
	scene = &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})
}

func initTextures() {
	index := 0
	textureFiles := map[int]string{
		HEAD: "head.jpg",
		BODY: "body.jpg",
		FOOD: "food.jpg"}
	for nodeType, path := range textureFiles {
		texture, err := app.Open(path)
		defer texture.Close()
		if err != nil {
			log.Fatal(err)
		}
		img, _, err := image.Decode(texture)
		if err != nil {
			log.Fatal(err)
		}
		tex, err := eng.LoadTexture(img)
		if err != nil {
			log.Fatal(err)
		}
		textures[nodeType] = sprite.SubTex{tex, img.Bounds()}
		index = index + 1
	}
}

type Node struct {
	NodeType   int
	SpriteNode *sprite.Node
	X          int
	Y          int
}

type SnakeNode struct {
	Value *Node
	Next  *SnakeNode
}

func newSnakeNode(nodeType int) *SnakeNode {
	return &SnakeNode{newNode(nodeType), nil}
}

func newNode(nodeType int) *Node {
	n := &sprite.Node{}
	eng.Register(n)
	scene.AppendChild(n)
	eng.SetSubTex(n, textures[nodeType])
	node := &Node{nodeType, n, 0, 0}
	node.Update()
	return node
}

func (node *Node) SetLocation(x, y int) {
	node.X = x
	node.Y = y
	node.Update()
}

func (node *Node) Update() {
	matrix := f32.Affine{
		{NODE_SIZE, 0, float32(node.X) * NODE_SIZE},
		{0, NODE_SIZE, float32(node.Y) * NODE_SIZE},
	}
	eng.SetTransform(node.SpriteNode, matrix)
}
