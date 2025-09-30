package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Lofanmi/chinese-calendar-golang/calendar"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/starainrt/astro/sun"
)

type Event struct {
	Name         string `mapstructure:"name"`
	Type         string `mapstructure:"type"`
	Year         int    `mapstructure:"year,omitempty"`
	Month        int    `mapstructure:"month"`
	Day          int    `mapstructure:"day"`
	RepeatYearly bool   `mapstructure:"repeatYearly"`
	AlwaysShow   bool   `mapstructure:"alwaysShow"`
	IsLunar      bool   `mapstructure:"isLunar"`
}

type Config struct {
	DatesFile              string `mapstructure:"datesFile"`
	SentencesFile          string `mapstructure:"sentencesFile"`
	ShowDateAmount         int    `mapstructure:"showDateAmount"`
	DateFormat             string `mapstructure:"dateFormat"`
	Location               Location `mapstructure:"location"`
	ShowSunTimes           bool   `mapstructure:"showSunTimes"`
	ShowDailySentence      bool   `mapstructure:"showDailySentence"`
	Events                 []Event `mapstructure:"events,omitempty"`
	CacheDir               string `mapstructure:"cacheDir"`
	SentenceUpdateMode     string `mapstructure:"sentenceUpdateMode"`
	SentenceUpdateInterval int    `mapstructure:"sentenceUpdateInterval"`
}

type Location struct {
	Latitude  float64 `mapstructure:"latitude"`
	Longitude float64 `mapstructure:"longitude"`
	Timezone  int     `mapstructure:"timezone"`
}

type EventWithDays struct {
	Event Event
	Days  int
	Date  time.Time
}

type SentenceCache struct {
	Sentence    string    `json:"sentence"`
	LastUpdate  time.Time `json:"lastUpdate"`
	UpdateCount int       `json:"updateCount"`
}

func main() {
	configDir := flag.String("config", ".", "配置文件目录")
	addSentence := flag.String("add-sentence", "", "添加新的文案到sentences.json")
	showSunTimes := flag.String("show-sun-times", "", "是否显示日出日落时间 (true/false)")
	showDailySentence := flag.String("show-daily-sentence", "", "是否显示每日文案 (true/false)")
	showDateAmount := flag.Int("show-date-amount", 0, "设置显示事件数量")
	addEvent := flag.Bool("add-event", false, "添加新事件（交互式）")
	refreshSentence := flag.Bool("refresh-sentence", false, "手动刷新文案")
	help := flag.Bool("h", false, "显示帮助信息")
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if _, err := os.Stat(*configDir); os.IsNotExist(err) {
		fmt.Printf("目录不存在: %s\n", *configDir)
		os.Exit(1)
	}

	if *addSentence != "" {
		handleAddSentence(*configDir, *addSentence)
		os.Exit(0)
	}

	if *addEvent {
		// 获取非标志参数作为事件名称
		eventName := ""
		if flag.NArg() > 0 {
			eventName = flag.Arg(0)
		}
		handleAddEvent(*configDir, eventName)
		os.Exit(0)
	}

	config := loadConfig(*configDir)
	handleConfigUpdates(*showSunTimes, *showDailySentence, *showDateAmount)

	displayInfo(*configDir, config, *refreshSentence)
}

func printHelp() {
	fmt.Println("终端时间显示工具 - 命令行参数:")
	fmt.Println("  -config string            配置文件目录 (默认 \".\")")
	fmt.Println("  -add-sentence string      添加新的文案到sentences.json")
	fmt.Println("  -show-sun-times string    是否显示日出日落时间 (true/false)")
	fmt.Println("  -show-daily-sentence string 是否显示每日文案 (true/false)")
	fmt.Println("  -show-date-amount int     设置显示事件数量")
	fmt.Println("  -add-event                添加新事件（交互式）")
	fmt.Println("  -refresh-sentence         手动刷新文案")
	fmt.Println("  -h                        显示帮助信息")
	fmt.Println("\n示例:")
	fmt.Println("  ./app -add-event              # 交互式添加事件（询问所有信息）")
	fmt.Println("  ./app -add-event 春节          # 指定名称，其余交互输入")
}

func handleAddSentence(configDir, sentence string) {
	sentencesPath := filepath.Join(configDir, "sentences.json")
	if err := addSentenceToFile(sentencesPath, sentence); err != nil {
		fmt.Printf("添加文案失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("成功添加文案: %s\n", sentence)
}

func loadConfig(configDir string) Config {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("读取主配置失败: %v\n", err)
		os.Exit(1)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("解析主配置失败: %v\n", err)
		os.Exit(1)
	}

	datesViper := viper.New()
	datesViper.SetConfigFile(filepath.Join(configDir, config.DatesFile))
	if err := datesViper.ReadInConfig(); err != nil {
		fmt.Printf("读取dates.json失败: %v\n", err)
		os.Exit(1)
	}

	var events []Event
	if err := datesViper.UnmarshalKey("events", &events); err != nil {
		fmt.Printf("解析events失败: %v\n", err)
		os.Exit(1)
	}
	config.Events = events

	setDefaults(&config)
	return config
}

func setDefaults(config *Config) {
	if config.ShowDateAmount == 0 {
		config.ShowDateAmount = 7
	}
	if config.DateFormat == "" {
		config.DateFormat = "2006/01/02"
	}
	if config.CacheDir == "" {
		config.CacheDir = "cache"
	}
}

func handleConfigUpdates(showSunTimes, showDailySentence string, showDateAmount int) {
	needSave := false

	if showSunTimes != "" {
		if val, err := strconv.ParseBool(showSunTimes); err == nil {
			viper.Set("showSunTimes", val)
			needSave = true
		}
	}

	if showDailySentence != "" {
		if val, err := strconv.ParseBool(showDailySentence); err == nil {
			viper.Set("showDailySentence", val)
			needSave = true
		}
	}

	if showDateAmount > 0 {
		viper.Set("showDateAmount", showDateAmount)
		needSave = true
	}

	if needSave {
		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleAddEvent(configDir, eventName string) {
	fmt.Println("\n=== 添加新事件 ===\n")
	
	// 读取配置以获取 datesFile 路径
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("读取配置失败: %v\n", err)
		os.Exit(1)
	}
	
	datesFile := viper.GetString("datesFile")
	if datesFile == "" {
		datesFile = "dates.json"
	}

	event := Event{}

	// 询问事件名称（如果未提供）
	if eventName == "" {
		fmt.Print("事件名称: ")
		fmt.Scanln(&event.Name)
		if event.Name == "" {
			// 如果 Scanln 失败，使用 bufio 读取整行
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				event.Name = strings.TrimSpace(scanner.Text())
			}
		}
		if event.Name == "" {
			fmt.Println("事件名称不能为空")
			os.Exit(1)
		}
	} else {
		event.Name = eventName
		fmt.Printf("事件名称: %s\n", event.Name)
	}

	// 事件类型
	fmt.Print("事件类型 (birthday/festival) [festival]: ")
	fmt.Scanln(&event.Type)
	if event.Type == "" {
		event.Type = "festival"
	}
	if event.Type != "birthday" && event.Type != "festival" {
		fmt.Println("无效的事件类型，已设置为 festival")
		event.Type = "festival"
	}

	// 历法类型
	var isLunarStr string
	fmt.Print("使用农历 (y/n) [n]: ")
	fmt.Scanln(&isLunarStr)
	event.IsLunar = strings.ToLower(isLunarStr) == "y"

	// 是否每年重复
	var repeatStr string
	fmt.Print("每年重复 (y/n) [y]: ")
	fmt.Scanln(&repeatStr)
	event.RepeatYearly = repeatStr == "" || strings.ToLower(repeatStr) == "y"

	// 输入日期
	if event.RepeatYearly {
		fmt.Print("请输入日期 (月 日，如: 1 1): ")
		fmt.Scanf("%d %d", &event.Month, &event.Day)
	} else {
		fmt.Print("请输入完整日期 (年 月 日，如: 2025 1 1): ")
		fmt.Scanf("%d %d %d", &event.Year, &event.Month, &event.Day)
	}

	// 验证日期
	if !validateDate(event) {
		fmt.Println("日期格式无效")
		os.Exit(1)
	}

	// 是否始终显示
	var alwaysShowStr string
	fmt.Print("始终显示 (y/n) [n]: ")
	fmt.Scanln(&alwaysShowStr)
	event.AlwaysShow = strings.ToLower(alwaysShowStr) == "y"

	// 显示确认信息
	fmt.Println("\n事件信息:")
	fmt.Printf("  名称: %s\n", event.Name)
	fmt.Printf("  类型: %s\n", event.Type)
	if event.IsLunar {
		fmt.Print("  历法: 农历\n")
	} else {
		fmt.Print("  历法: 公历\n")
	}
	if event.RepeatYearly {
		fmt.Printf("  日期: 每年 %d月%d日\n", event.Month, event.Day)
	} else {
		fmt.Printf("  日期: %d年%d月%d日\n", event.Year, event.Month, event.Day)
	}
	fmt.Printf("  始终显示: %v\n", event.AlwaysShow)

	var confirm string
	fmt.Print("\n确认添加? (y/n) [y]: ")
	fmt.Scanln(&confirm)
	if confirm != "" && strings.ToLower(confirm) != "y" {
		fmt.Println("已取消")
		os.Exit(0)
	}

	// 添加到 dates.json
	if err := addEventToFile(configDir, datesFile, event); err != nil {
		fmt.Printf("添加事件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ 成功添加事件: %s\n", event.Name)
}

func validateDate(event Event) bool {
	if event.Month < 1 || event.Month > 12 {
		return false
	}
	if event.Day < 1 || event.Day > 31 {
		return false
	}
	if !event.RepeatYearly && event.Year < 1900 {
		return false
	}
	return true
}

func addEventToFile(configDir, datesFile string, event Event) error {
	filePath := filepath.Join(configDir, datesFile)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var eventsData struct {
		Events []Event `json:"events"`
	}
	if err := json.Unmarshal(data, &eventsData); err != nil {
		return err
	}

	eventsData.Events = append(eventsData.Events, event)

	newData, err := json.MarshalIndent(eventsData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, newData, 0644)
}

func displayInfo(configDir string, config Config, refresh bool) {
	now := time.Now()

	color.New(color.FgCyan, color.Bold).Println("Hello World! \n")

	if config.ShowDailySentence {
		if sentence, err := getDailySentence(configDir, config, refresh); err == nil {
			color.New(color.FgMagenta).Printf("  📜 %s\n\n", sentence)
		}
	}

	if config.ShowSunTimes && config.Location.Latitude != 0 && config.Location.Longitude != 0 {
		displaySunTimes(config.Location, now)
	}

	displayEvents(config, now)
	fmt.Println()
}

func displaySunTimes(location Location, now time.Time) {
	sunrise, sunset, err := calculateSunTimes(location, now)
	if err != nil {
		return
	}

	sunriseTime, sunriseIsNext := getNextSunTime(sunrise, now, location, true)
	sunsetTime, sunsetIsNext := getNextSunTime(sunset, now, location, false)

	sunriseDiff := sunriseTime.Sub(now)
	sunsetDiff := sunsetTime.Sub(now)

	sunColor := color.New(color.FgYellow)
	if sunriseDiff < sunsetDiff {
		prefix := "这次"
		if sunriseIsNext {
			prefix = "下次"
		}
		sunColor.Printf("  ☀️ 距离%s日出还有 %d小时%d分钟\n", prefix, int(sunriseDiff.Hours()), int(sunriseDiff.Minutes())%60)
	} else {
		prefix := "这次"
		if sunsetIsNext {
			prefix = "下次"
		}
		sunColor.Printf("  🌙 距离%s日落还有 %d小时%d分钟\n", prefix, int(sunsetDiff.Hours()), int(sunsetDiff.Minutes())%60)
	}
}

func getNextSunTime(sunTime time.Time, now time.Time, location Location, isSunrise bool) (time.Time, bool) {
	if sunTime.After(now) {
		return sunTime, false
	}

	tomorrow := now.Add(24 * time.Hour)
	sunrise, sunset, _ := calculateSunTimes(location, tomorrow)
	if isSunrise {
		return sunrise, true
	}
	return sunset, true
}

func calculateSunTimes(location Location, date time.Time) (sunrise, sunset time.Time, err error) {
	loc := time.FixedZone("Custom", location.Timezone*3600)
	date = date.In(loc)

	sunrise, err = sun.RiseTime(date, location.Longitude, location.Latitude, 0, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sunrise = sunrise.In(loc)

	sunset, err = sun.SetTime(date, location.Longitude, location.Latitude, 0, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	sunset = sunset.In(loc)

	return sunrise, sunset, nil
}

func displayEvents(config Config, now time.Time) {
	events := collectEvents(config.Events, now)
	filteredEvents := filterEvents(events, config.ShowDateAmount)

	for _, e := range filteredEvents {
		emoji := "🎉"
		displayColor := color.New(color.FgBlue)
		if e.Event.Type == "birthday" {
			emoji = "🎂"
			displayColor = color.New(color.FgGreen)
		}

		if e.Days >= 0 {
			displayColor.Printf("  %s %s 还有 %d 天 (%s)\n", emoji, e.Event.Name, e.Days, e.Date.Format(config.DateFormat))
		} else {
			displayColor.Printf("  %s %s 已经过去 %d 天 (%s)\n", emoji, e.Event.Name, -e.Days, e.Date.Format(config.DateFormat))
		}
	}
}

func collectEvents(events []Event, now time.Time) []EventWithDays {
	var result []EventWithDays
	for _, e := range events {
		target := calculateTargetDate(e, now)
		days := int(target.Sub(now).Hours()) / 24
		result = append(result, EventWithDays{e, days, target})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Days < result[j].Days
	})

	return result
}

func calculateTargetDate(event Event, now time.Time) time.Time {
	if event.IsLunar {
		lunarDate := convertLunarToSolar(now.Year(), event.Month, event.Day)
		target := time.Date(lunarDate.Year, time.Month(lunarDate.Month), lunarDate.Day, 0, 0, 0, 0, time.Local)
		
		if event.RepeatYearly && target.Before(now) {
			nextYearLunar := convertLunarToSolar(now.Year()+1, event.Month, event.Day)
			target = time.Date(nextYearLunar.Year, time.Month(nextYearLunar.Month), nextYearLunar.Day, 0, 0, 0, 0, time.Local)
		}
		return target
	}

	target := time.Date(now.Year(), time.Month(event.Month), event.Day, 0, 0, 0, 0, time.Local)
	if event.RepeatYearly && target.Before(now) {
		target = time.Date(now.Year()+1, time.Month(event.Month), event.Day, 0, 0, 0, 0, time.Local)
	} else if !event.RepeatYearly {
		target = time.Date(event.Year, time.Month(event.Month), event.Day, 0, 0, 0, 0, time.Local)
	}
	
	return target
}

func filterEvents(allEvents []EventWithDays, maxCount int) []EventWithDays {
	var filtered []EventWithDays
	
	for _, e := range allEvents {
		if e.Event.AlwaysShow {
			filtered = append(filtered, e)
		}
	}

	for _, e := range allEvents {
		if !e.Event.AlwaysShow && len(filtered) < maxCount {
			filtered = append(filtered, e)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Days < filtered[j].Days
	})

	return filtered
}

func getDailySentence(configDir string, config Config, refresh bool) (string, error) {
	cachePath := filepath.Join(configDir, config.CacheDir, "sentence.cache")
	
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return "", err
	}

	var cache SentenceCache
	if !refresh {
		if sentence, ok := tryLoadCache(cachePath, &cache, config); ok {
			return sentence, nil
		}
	}

	sentences, err := loadSentences(filepath.Join(configDir, config.SentencesFile))
	if err != nil {
		return "", err
	}

	sentence := selectRandomSentence(sentences)
	cache.Sentence = sentence
	cache.LastUpdate = time.Now()
	
	if refresh || config.SentenceUpdateMode == "count" {
		cache.UpdateCount = 0
	}

	saveCache(cachePath, &cache)
	return sentence, nil
}

func tryLoadCache(cachePath string, cache *SentenceCache, config Config) (string, bool) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return "", false
	}

	if err := json.Unmarshal(data, cache); err != nil {
		return "", false
	}

	if needUpdate(cache, config) {
		return "", false
	}

	if config.SentenceUpdateMode == "count" {
		cache.UpdateCount++
		saveCache(cachePath, cache)
	}

	return cache.Sentence, true
}

func needUpdate(cache *SentenceCache, config Config) bool {
	if config.SentenceUpdateMode == "time" {
		return time.Since(cache.LastUpdate) > time.Duration(config.SentenceUpdateInterval)*time.Hour
	}
	if config.SentenceUpdateMode == "count" {
		return cache.UpdateCount >= config.SentenceUpdateInterval
	}
	return false
}

func saveCache(cachePath string, cache *SentenceCache) {
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	tempPath := cachePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return
	}
	os.Rename(tempPath, cachePath)
}

func loadSentences(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var sentencesData struct {
		Sentences []string `json:"sentences"`
	}
	if err := json.Unmarshal(data, &sentencesData); err != nil {
		return nil, err
	}

	if len(sentencesData.Sentences) == 0 {
		return nil, fmt.Errorf("sentences.json为空")
	}

	return sentencesData.Sentences, nil
}

func selectRandomSentence(sentences []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return sentences[r.Intn(len(sentences))]
}

func addSentenceToFile(filePath, newSentence string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var sentencesData struct {
		Sentences []string `json:"sentences"`
	}
	if err := json.Unmarshal(data, &sentencesData); err != nil {
		return err
	}

	sentencesData.Sentences = append(sentencesData.Sentences, newSentence)

	newData, err := json.MarshalIndent(sentencesData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, newData, 0644)
}

func convertLunarToSolar(year, month, day int) struct{ Year, Month, Day int } {
	c := calendar.ByLunar(int64(year), int64(month), int64(day), 0, 0, 0, false)
	return struct{ Year, Month, Day int }{
		Year:  int(c.Solar.GetYear()),
		Month: int(c.Solar.GetMonth()),
		Day:   int(c.Solar.GetDay()),
	}
}