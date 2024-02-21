package main

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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	initialRootNumNodes = 10
	numEachLayer        = 5
	numLayers           = 9
	numNodes            = 3000
	maxLayer            = 200
	inviteeNumArr       = []int{}
	id                  = 0
	inviteeNum          = []int{10, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	invitorsProbability = []float64{10, 10, 7, 1, 1, 1, 1, 1, 1, 1}
	depositNum          = []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	depositProbability  = []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	// Uncomment the following line if your bare application
	// has an action associated with it:

	Run: func(cmd *cobra.Command, args []string) {
		initialRootNumNodes = viper.GetInt("initialRootNumNodes")
		numEachLayer = viper.GetInt("numEachLayer")
		numLayers = viper.GetInt("numLayers")
		numNodes = viper.GetInt("numNodes")
		inviteeNum = viper.GetIntSlice("inviteeNum")
		viper.UnmarshalKey("invitorsProbability", &invitorsProbability)
		viper.UnmarshalKey("depositProbability", &depositProbability)

		initInviteeNum()
		initDeposit()
		initLayerNodes()
		sumDeposit()
		toCSV()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

var CfgFile string

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		// Search config in home directory with name .config" (without extension).
		// 获取当前文件的路径
		currentPath, err := os.Executable()
		if err != nil {
			log.Fatalln(err)
		}
		currentDir := filepath.Dir(currentPath)
		viper.AddConfigPath(currentDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml") //设置配置文件类型，可选
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Using config file err:", err)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	//The default should be the production environment.
	viper.SetDefault("environment", "production")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type Node struct {
	ID             int
	Layer          int
	InviteeNum     int
	RealInviteeNum int
	ParentID       int
	InviteeID      int
	Path           []int
	Deopsit        int
}

var userLayerMap = make(map[int]int)

func newNode(layer, parentID, inviteeID int, path []int) (*Node, error) {
	id, err := getID()
	if err != nil {
		return nil, err
	}
	userLayerMap[id] = layer
	return &Node{
		ID:             id,
		Layer:          layer,
		InviteeNum:     getInviteeNum(),
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
		node, err := newNode(0, 0, 0, []int{})
		if err != nil {
			log.Println(err)
			return
		}
		maxLayerMap[0] = append(maxLayerMap[0], node)
	}
	inviteLayerNodes(1)
}

// 按照层级邀请节点
func inviteLayerNodes(layer int) {
	println("invite.layer", layer)
	if layer > maxLayer {
		return
	}
	for _, node := range maxLayerMap[layer-1] {
		err := inviteMapNode(node)
		if err != nil {
			log.Println(err)
			return
		}
	}
	inviteLayerNodes(layer + 1)
}

// 邀请节点
func inviteMapNode(node *Node) error {
	inviteeNum := node.InviteeNum
	for i := 0; i < inviteeNum; i++ {
		parentNode := parentLayer(node)
		layer := parentNode.Layer + 1
		childNode, err := newNode(layer, parentNode.ID, node.ID, append(parentNode.Path, parentNode.ID))
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
	filename := fmt.Sprintf("%s/data-%s.csv", currentDir, time.Now().Format("2006-01-02-15-04-05"))
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
		"id", "layer", "invitee_num", "real_invitee_num", "parent_id", "invitee_id", "deopsit", "path",
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
