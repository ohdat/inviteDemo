package invite

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	initialRootNumNodes     = 1000
	numEachLayer            = 5
	numLayers               = 9
	numNodes                = 3000
	maxLayer                = 20
	inviteeNumArr           = []int{}
	RootInviteeNumArr       = []int{}
	id                      = 0
	randomInvite            = false
	inviteeNum              = []int{10, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	invitorsProbability     = []float64{10, 10, 7, 1, 1, 1, 1, 1, 1, 1}
	depositNum              = []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	depositProbability      = []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	Day                     = 10
	DayProbability          = []float64{40, 30, 20, 10}
	RootInviteeNum          = []int{10, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	RootInvitorsProbability = []float64{10, 10, 7, 1, 1, 1, 1, 1, 1, 1}
)

func Invite() {
	initialRootNumNodes = viper.GetInt("initialRootNumNodes")
	numEachLayer = viper.GetInt("numEachLayer")
	numLayers = viper.GetInt("numLayers")
	numNodes = viper.GetInt("numNodes")
	inviteeNum = viper.GetIntSlice("inviteeNum")
	depositNum = viper.GetIntSlice("depositNum")
	randomInvite = viper.GetBool("randomInvite")
	RootInviteeNum = viper.GetIntSlice("rootInviteeNum")
	Day = viper.GetInt("day")
	viper.UnmarshalKey("invitorsProbability", &invitorsProbability)
	viper.UnmarshalKey("rootInvitorsProbability", &RootInvitorsProbability)
	viper.UnmarshalKey("depositProbability", &depositProbability)
	viper.UnmarshalKey("dayProbability", &DayProbability)

	initInviteeNum()
	initDeposit()
	initLayerNodes()
	sumDeposit()
	toCSV()
}

type Node struct {
	ID             int
	Layer          int
	Day            int
	InviteeNum     int
	RealInviteeNum int
	ParentID       int
	InviteeID      int
	Path           []int
	Deopsit        int
}

var userLayerMap = make(map[int]int)

func newNode(isRoot bool, day, layer, parentID, inviteeID int, path []int) (*Node, error) {
	id, err := getID()
	if err != nil {
		return nil, err
	}
	var inviteeNum = 0
	if isRoot {
		inviteeNum = getRootInviteeNum()
	} else {
		inviteeNum = getInviteeNum()
	}
	userLayerMap[id] = layer
	return &Node{
		ID:             id,
		Layer:          layer,
		Day:            day,
		InviteeNum:     inviteeNum,
		RealInviteeNum: 0,
		ParentID:       parentID,
		InviteeID:      inviteeID,
		Deopsit:        getDepositNum(),
		Path:           path,
	}, nil
}

var (
	maxLayerMap = make(map[int][]*Node)
)

// 初始化第一层节点
func initLayerNodes() {
	for i := 0; i < initialRootNumNodes; i++ {
		node, err := newNode(true, 0, 0, 0, 0, []int{})
		if err != nil {
			log.Println(err)
			return
		}
		maxLayerMap[0] = append(maxLayerMap[0], node)
	}
	for i := 1; i < maxLayer; i++ {
		var ifCont = false
		for j := 0; j < Day; j++ {
			err := inviteLayerNodes(i, j+1)
			if err != nil {
				return
				ifCont = true
				break
			}
		}
		if ifCont {
			continue
		}
	}
}

// 按照层级邀请节点
func inviteLayerNodes(layer int, day int) error {
	println("invite.layer", layer, "day", day)
	for _, node := range maxLayerMap[layer-1] {
		err := inviteMapNode(node, day)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
func getInviteeNumByDay(day int, intieeNum int) int {
	if day < 1 {
		return 0
	}
	if len(DayProbability) < day {
		return 0
	}
	dayProbability := DayProbability[day-1]
	return int(intieeNum * int(dayProbability) / 100)
}

// 邀请节点
func inviteMapNode(node *Node, day int) error {
	inviteeNum := getInviteeNumByDay(day-node.Day, node.InviteeNum)
	for i := 0; i < inviteeNum; i++ {
		parentNode := parentLayer(node)
		layer := parentNode.Layer + 1
		childNode, err := newNode(false, day, layer, parentNode.ID, node.ID, append(parentNode.Path, parentNode.ID))
		if err != nil {
			log.Println(err)
			return err
		}
		// log.Println("node.ID", node.ID)
		node.RealInviteeNum = node.RealInviteeNum + 1
		// addMapInviteeNum(node)
		maxLayerMap[layer] = append(maxLayerMap[layer], childNode)
	}
	return nil
}

// 找到要添加的父节点
func parentLayer(node *Node) *Node {
	layer := node.Layer + 1
	var childNum = 0
	for _, childNode := range maxLayerMap[layer] {
		if childNode.ParentID == node.ID {
			childNum++
		}
	}
	if childNum >= numEachLayer {
		for _, childNode := range maxLayerMap[layer] {
			if childNode.ParentID == node.ID {
				return parentLayer(childNode)
			}
		}
	}
	return node
}

//new =================================

func getInviteeNum() int {
	if len(inviteeNumArr) == 0 {
		return 0
	}
	if randomInvite {
		//随机从inviteeNumArr 中取出一个数
		randomNum := rand.Intn(len(inviteeNumArr))
		inviteeNum := inviteeNumArr[randomNum]
		inviteeNumArr = append(inviteeNumArr[:randomNum], inviteeNumArr[randomNum+1:]...)
		return inviteeNum
	}

	inviteeNum := inviteeNumArr[0]
	inviteeNumArr = inviteeNumArr[1:]
	return inviteeNum
}

func initInviteeNum() {

	var ratio float64 = 0
	for i := range inviteeNum {
		ratio += float64(inviteeNum[i]) * invitorsProbability[i]
	}
	log.Println("ratio:", ratio)
	numBase := float64(numNodes) / 100
	newTotal := 0
	for i := range inviteeNum {
		for j := 0; j < int(invitorsProbability[i]*numBase); j++ {
			newTotal += inviteeNum[i]
			inviteeNumArr = append(inviteeNumArr, inviteeNum[i])
		}
	}
	log.Println("newTotal:", newTotal)
	log.Println("len(inviteeNumArr):", len(inviteeNumArr))
	initRootInviteeNum()
}

func getRootInviteeNum() int {
	if len(RootInviteeNumArr) == 0 {
		return 0
	}
	if randomInvite {
		//随机从RootInviteeNumArr 中取出一个数
		randomNum := rand.Intn(len(RootInviteeNumArr))
		inviteeNum := RootInviteeNumArr[randomNum]
		RootInviteeNumArr = append(RootInviteeNumArr[:randomNum], RootInviteeNumArr[randomNum+1:]...)
		return inviteeNum
	}

	inviteeNum := RootInviteeNum[0]
	RootInviteeNumArr = RootInviteeNumArr[1:]
	return inviteeNum
}

func initRootInviteeNum() {
	for i := range RootInviteeNum {
		for j := 0; j < int(RootInvitorsProbability[i]*float64(initialRootNumNodes)/100); j++ {
			RootInviteeNumArr = append(RootInviteeNumArr, RootInviteeNum[i])
		}
	}
}

func getID() (int, error) {
	if id > numNodes {
		err := fmt.Errorf("id > numNodes", id)
		return 0, err
	}
	id++
	return id, nil
}

func toCSV() {
	currentPath, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	currentDir := filepath.Dir(currentPath)
	filename := fmt.Sprintf("%s/data-%s.csv", currentDir, time.Now().Format("2006-01-02-15-04"))
	if viper.GetString("INVITEDEV") == "1" {
		filename = fmt.Sprintf("./data-%s.csv", time.Now().Format("2006-01-02-15-04"))
	}
	fmt.Println("filename:", filename)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Cannot create file", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	tableTitle := []string{
		"id", "day", "layer", "invitee_num", "real_invitee_num", "parent_id", "invitee_id", "deopsit", "path",
	}
	for i := 0; i < numLayers; i++ {
		tableTitle = append(tableTitle, "sum_deposit_"+strconv.Itoa(i+1))
	}
	writer.Write(tableTitle)
	fmt.Printf("共%d个用户数据\n", id-1)
	newtoData(writer)
}

var itodat = 0

func newtoData(writer *csv.Writer) {
	// println("itodat:", itodat)
	for _, nodes := range maxLayerMap {
		for _, node := range nodes {
			itodat++
			row := []string{
				strconv.Itoa(node.ID),
				strconv.Itoa(node.Day),
				strconv.Itoa(node.Layer),
				strconv.Itoa(node.InviteeNum),
				strconv.Itoa(node.RealInviteeNum),
				strconv.Itoa(node.ParentID),
				strconv.Itoa(node.InviteeID),
				strconv.Itoa(node.Deopsit),
				strings.Trim(strings.Join(strings.Fields(fmt.Sprint(node.Path)), "_"), "[]"),
			}
			sum := sumDepositMap[node.ID]
			for i := 0; i < numLayers; i++ {
				if i < len(sum) {
					deposit := sum[i]
					row = append(row, strconv.Itoa(deposit))
				}
			}
			writer.Write(row)
		}
	}
}

var depositNumArr = []int{}

// deposit
func initDeposit() {
	var total = numNodes
	for i := range depositNum {
		for j := 0; j < int(depositProbability[i]*float64(total)/100); j++ {
			depositNumArr = append(depositNumArr, depositNum[i])
		}
	}
	log.Println("len(depositNumArr):", len(depositNumArr))
}
func getDepositNum() int {
	if len(depositNumArr) == 0 {
		return 0
	}
	//随机从depositNumArr 中取出一个数
	randomNum := rand.Intn(len(depositNumArr))
	depositNum := depositNumArr[randomNum]
	depositNumArr = append(depositNumArr[:randomNum], depositNumArr[randomNum+1:]...)
	return depositNum
}

var sumDepositMap map[int][]int

// sum deposit
func sumDeposit() {
	sumDepositMap = make(map[int][]int)
	for _, nodes := range maxLayerMap {
		for _, node := range nodes {
			if node.Layer >= 0 {
				nodePath := node.Path
				if len(nodePath) > numLayers {
					nodePath = nodePath[len(nodePath)-numLayers:]
				}
				for _, id := range nodePath {
					addDepositMap(id, node.Deopsit, node.Layer)
					// layersNum := u
					// println("layersNum", sumDepositMap[id])
					// sumDepositMap[id][layersNum-1] += node.Deopsit
				}
			}
		}
	}
}

// addDeposit
func addDepositMap(id int, deposit int, layer int) {
	layerIndex := layer - userLayerMap[id] - 1
	if sumDepositMap[id] == nil {
		sumDepositMap[id] = make([]int, numLayers)
	}
	sumDepositMap[id][layerIndex] += deposit
}
