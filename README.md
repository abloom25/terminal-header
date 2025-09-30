# 终端时间显示工具

一个功能丰富的终端日期提醒工具，支持公历/农历事件管理、日出日落时间显示和每日文案功能。

## 功能特性

- 📅 **事件管理**：支持生日、节日等重要日期提醒
- 🌙 **农历支持**：完美支持农历日期转换
- ☀️ **日出日落**：根据地理位置显示日出日落时间
- 📜 **每日文案**：随机展示激励文案，支持时间/次数模式更新
- 🎨 **彩色输出**：美观的终端彩色显示
- ⚙️ **灵活配置**：JSON 配置文件，易于管理

## 安装

```bash
# 克隆仓库
git clone <repository-url>
cd <project-directory>

# 安装依赖
go mod tidy

# 编译
go build -o app main.go
```

## 配置文件

### config.json（主配置文件）

```json
{
  "datesFile": "dates.json",
  "sentencesFile": "sentences.json",
  "showDateAmount": 7,
  "dateFormat": "2006/01/02",
  "location": {
    "latitude": 39.9042,
    "longitude": 116.4074,
    "timezone": 8
  },
  "showSunTimes": true,
  "showDailySentence": true,
  "cacheDir": "cache",
  "sentenceUpdateMode": "count",
  "sentenceUpdateInterval": 5
}
```

**配置说明：**
- `datesFile`: 事件数据文件路径
- `sentencesFile`: 文案数据文件路径
- `showDateAmount`: 显示最近的事件数量（默认 7）
- `dateFormat`: 日期显示格式
- `location`: 地理位置信息
  - `latitude`: 纬度
  - `longitude`: 经度
  - `timezone`: 时区（UTC 偏移小时数）
- `showSunTimes`: 是否显示日出日落时间
- `showDailySentence`: 是否显示每日文案
- `cacheDir`: 缓存目录
- `sentenceUpdateMode`: 文案更新模式
  - `time`: 基于时间更新（单位：小时）
  - `count`: 基于运行次数更新
- `sentenceUpdateInterval`: 更新间隔值

### dates.json（事件数据）

```json
{
  "events": [
    {
      "name": "春节",
      "type": "festival",
      "month": 1,
      "day": 1,
      "repeatYearly": true,
      "alwaysShow": true,
      "isLunar": true
    },
    {
      "name": "小明生日",
      "type": "birthday",
      "year": 2025,
      "month": 5,
      "day": 20,
      "repeatYearly": true,
      "alwaysShow": false,
      "isLunar": false
    }
  ]
}
```

**字段说明：**
- `name`: 事件名称
- `type`: 事件类型（`birthday` 或 `festival`）
- `year`: 年份（仅在不重复时需要）
- `month`: 月份
- `day`: 日期
- `repeatYearly`: 是否每年重复
- `alwaysShow`: 是否始终显示（不受 `showDateAmount` 限制）
- `isLunar`: 是否使用农历

### sentences.json（文案数据）

```json
{
  "sentences": [
    "今天也要加油哦！",
    "生活明朗，万物可爱。",
    "星光不问赶路人，时光不负有心人。",
    "愿你走出半生，归来仍是少年。"
  ]
}
```

## 使用方法

### 基本运行

```bash
# 使用当前目录的配置文件
./app

# 指定配置目录
./app -config /path/to/config
```

### 添加事件

**交互式添加（推荐）：**

```bash
# 完全交互式
./app -add-event

# 指定事件名称，其余交互输入
./app -add-event 春节
```

交互流程示例：
```
=== 添加新事件 ===

事件名称: 春节
事件类型 (birthday/festival) [festival]: festival
使用农历 (y/n) [n]: y
每年重复 (y/n) [y]: y
请输入日期 (月 日，如: 1 1): 1 1
始终显示 (y/n) [n]: y

事件信息:
  名称: 春节
  类型: festival
  历法: 农历
  日期: 每年 1月1日
  始终显示: true

确认添加? (y/n) [y]: y

✓ 成功添加事件: 春节
```

### 添加文案

```bash
./app -add-sentence "你的新文案内容"
```

### 刷新文案

```bash
# 手动刷新当前显示的文案
./app -refresh-sentence
```

### 配置修改

```bash
# 开启/关闭日出日落显示
./app -show-sun-times true
./app -show-sun-times false

# 开启/关闭每日文案
./app -show-daily-sentence true
./app -show-daily-sentence false

# 设置显示事件数量
./app -show-date-amount 10
```

### 查看帮助

```bash
./app -h
```

## 输出示例

```
Hello World!

  📜 今天也要加油哦！

  ☀️ 距离这次日落还有 3小时45分钟
  🎂 小明生日 还有 15 天 (2025/05/20)
  🎉 春节 还有 120 天 (2026/01/29)
  🎉 中秋节 还有 200 天 (2025/10/06)
```

## 文案更新模式

### 时间模式（time）

根据时间间隔自动更新文案：

```json
{
  "sentenceUpdateMode": "time",
  "sentenceUpdateInterval": 24
}
```

上述配置表示每 24 小时更新一次文案。

### 计数模式（count）

根据运行次数更新文案：

```json
{
  "sentenceUpdateMode": "count",
  "sentenceUpdateInterval": 5
}
```

上述配置表示每运行 5 次程序更新一次文案。

## 事件显示逻辑

1. 标记为 `alwaysShow: true` 的事件始终显示
2. 其余事件按时间从近到远排序
3. 最多显示 `showDateAmount` 个事件
4. 已过去的事件显示负数天数

## 日出日落计算

- 根据配置的经纬度和时区自动计算
- 显示距离最近的日出或日落时间
- 自动判断"这次"或"下次"

## 依赖项

```go
require (
    github.com/Lofanmi/chinese-calendar-golang/calendar
    github.com/fatih/color
    github.com/spf13/viper
    github.com/starainrt/astro/sun
)
```

## 常见问题

**Q: 如何设置开机自动运行？**

A: 可以将程序添加到系统的启动脚本或使用定时任务。

Linux/Mac (添加到 ~/.bashrc 或 ~/.zshrc):
```bash
/path/to/app -config /path/to/config
```

Windows (添加到启动文件夹或任务计划程序):
```
C:\path\to\app.exe -config C:\path\to\config
```

**Q: 文案一直不更新？**

A: 
1. 检查 `sentenceUpdateMode` 配置是否正确
2. 计数模式下需要运行足够次数
3. 可以使用 `-refresh-sentence` 手动刷新

**Q: 农历日期不准确？**

A: 程序使用专业的农历转换库，如有问题请检查日期输入格式是否正确。

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！

## 作者

ABloom25
