package holiday

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/thoas/go-funk"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dateFormat      = "2006-01-02"
	timeFormat      = "2006-01-02 15:04:05"
	fileSp          = "---------------"
	defaultFilename = "holiday-data"
)

// 固定节假日名称
var legalHolidays = []string{
	"元旦",
	"春节",
	"清明节",
	"劳动节",
	"端午节",
	"中秋节",
	"国庆节",
}

// 参数
type Config struct {
	Filename string // 数据文件名
}

type ChinaHoliday struct {
	f        bool   // 文件存储
	filename string // 文件名
}

// 创建实例
func New(config *Config) (*ChinaHoliday, error) {
	if config == nil {
		config = &Config{
			Filename: defaultFilename,
		}
	}
	f := true
	if config.Filename == "" {
		f = false
	}
	if f {
		filename := config.Filename
		info, err := os.Stat(filename)
		var file *os.File
		if err != nil {
			// 创建文件
			file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return nil, err
			}
		} else if info.IsDir() {
			return nil, errors.New(fmt.Sprintf("文件名%s不能是目录", filename))
		}
		// 及时关闭file句柄
		defer file.Close()
	}
	ins := ChinaHoliday{
		f:        f,
		filename: config.Filename,
	}
	return &ins, nil
}

// 检查某个日期是否是节假日(支持yyyy-MM-dd)
func (s *ChinaHoliday) Check(date string) (bool, error) {
	// 日期转换
	now, err := time.ParseInLocation(dateFormat, date, time.Local)
	if err != nil {
		return false, err
	}

	year := now.Year()
	holidays, workdays, err := s.all(year)
	if err != nil {
		return false, err
	}

	// 校验是否节假日
	nowDate := now.Format(dateFormat)
	if funk.ContainsString(holidays, nowDate) {
		// 节假日
		return true, nil
	} else {
		weekDay := now.Weekday()
		if weekDay == time.Sunday || weekDay == time.Saturday {
			if !funk.ContainsString(workdays, nowDate) {
				// 非调休日
				return true, nil
			}
		}
	}
	return false, nil
}

// 获取某一年节假日数目(周末除外)
func (s *ChinaHoliday) List(year int) ([]string, []string, error) {
	return s.all(year)
}

// 获取指定日期之间有多少节假日
func (s *ChinaHoliday) Range(startDate, endDate string) ([]string, []string, error) {
	years := getYears(startDate, endDate)
	allHolidays, allWorkdays, err := s.all(years...)
	if err != nil {
		return nil, nil, err
	}
	holidays := make([]string, 0)
	workdays := make([]string, 0)
	for _, holiday := range allHolidays {
		_, _, after1 := timeAfter(startDate, holiday)
		_, _, after2 := timeAfter(holiday, endDate)
		if after1 && after2 {
			holidays = append(holidays, holiday)
		}
	}
	for _, workday := range allWorkdays {
		_, _, after1 := timeAfter(startDate, workday)
		_, _, after2 := timeAfter(workday, endDate)
		if after1 && after2 {
			workdays = append(workdays, workday)
		}
	}
	return holidays, workdays, nil
}

// 读取全部节假日(一年或多年)
func (s *ChinaHoliday) all(years ...int) ([]string, []string, error) {
	holidays := make([]string, 0)
	workdays := make([]string, 0)
	// 从文件中读取数据
	list, err := s.getFromFile()
	for _, year := range years {
		currentHolidays := make([]string, 0)
		currentWorkdays := make([]string, 0)
		if err == nil {
			if oldData, ok := list[year]; ok {
				if len(oldData) == 1 {
					currentHolidays = oldData[0]
				} else if len(oldData) == 2 {
					currentHolidays = oldData[0]
					currentWorkdays = oldData[1]
				}
			}
		}
		if len(currentHolidays) == 0 && len(currentWorkdays) == 0 {
			// 读取线上数据
			currentHolidays, currentWorkdays, err = s.online(year)
			if err == nil && s.f {
				// 存入文件
				err = s.appendToFile(year, currentHolidays, currentWorkdays)
				if err != nil {
					return nil, nil, err
				}
			} else if err != nil {
				return nil, nil, err
			}
		}
		if len(currentHolidays) > 0 {
			holidays = append(holidays, currentHolidays...)
		}
		if len(currentWorkdays) > 0 {
			workdays = append(workdays, currentWorkdays...)
		}
	}
	return holidays, workdays, nil
}

// 从文件中获取数据
func (s *ChinaHoliday) getFromFile() (map[int][][]string, error) {
	list := make(map[int][][]string)
	if !s.f {
		return list, nil
	}
	file, err := os.OpenFile(s.filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	br := bufio.NewReader(file)
	year := 0
	holidays := make([]string, 0)
	workdays := make([]string, 0)
	i := 0
	for {
		// 按行读取
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		line := string(a)
		if line == fileSp {
			item := make([][]string, 0)
			item = append(item, holidays)
			item = append(item, workdays)
			list[year] = item
			i = 0
			continue
		} else if i == 0 {
			year = str2Int(line)
		} else if i == 1 {
			holidays = strings.Split(line, ",")
		} else if i == 2 {
			workdays = strings.Split(line, ",")
		}
		i++
	}
	return list, nil
}

// 写入文件
func (s *ChinaHoliday) appendToFile(year int, holidays, workdays []string) error {
	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入数据文件
	write := bufio.NewWriter(file)
	write.WriteString(fmt.Sprintf("%d\n", year))
	write.WriteString(fmt.Sprintf("%s\n", strings.Join(holidays, ",")))
	write.WriteString(fmt.Sprintf("%s\n", strings.Join(workdays, ",")))
	write.WriteString(fmt.Sprintf("%s\n", fileSp))
	write.Flush()
	return nil
}

// 刷新文件
func (s *ChinaHoliday) refreshFile(year int, holidays, workdays []string) error {
	b, err := ioutil.ReadFile(s.filename)
	if err != nil {
		return err
	}
	arr := strings.Split(strings.Trim(string(b), "\n"), fmt.Sprintf("%s\n", fileSp))
	newArr := make([]string, 0)
	for _, item := range arr {
		if item == "" {
			continue
		}
		itemArr := strings.Split(item, "\n")
		if len(itemArr) > 0 {
			if str2Int(itemArr[0]) == year {
				continue
			}
		}
		newArr = append(newArr, item)
	}
	newArr = append(newArr,
		fmt.Sprintf(
			"%d\n%s\n%s\n%s\n",
			year,
			strings.Join(holidays, ","),
			strings.Join(workdays, ","),
			// 最后一条记录需要加分隔符
			fileSp,
		),
	)
	// 覆盖旧文件
	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString(strings.Join(newArr, fmt.Sprintf("%s\n", fileSp)))
	write.Flush()
	return nil
}

// 或取指定年份的线上节假日数据
func (s *ChinaHoliday) online(year int) ([]string, []string, error) {
	c := colly.NewCollector(
		// 允许重复访问
		colly.AllowURLRevisit(),
	)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	q := fmt.Sprintf("国务院办公厅关于%d年部分节假日安排的通知", year)
	// url加密, 避免中文无法识别
	u := url.Values{}
	u.Set("t", "paper")
	u.Set("advance", "false")
	u.Set("q", q)
	searchUrl := "http://sousuo.gov.cn/s.htm?" + u.Encode()

	// 休假时间
	holidays := make([]string, 0)
	// 调休时间
	workdays := make([]string, 0)
	// 上一年调休时间(GOV一般是11月份发布第二年, 可能涉及跨年调休, 该调休应该属于上一年)
	lastYearWorkdays := make([]string, 0)

	// 查看所有a标签
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if e.Request.URL.String() == searchUrl {
			link := e.Attr("href")
			text := e.Text
			// 链接标题与搜索关键字一致
			if text == q {
				c.Visit(e.Request.AbsoluteURL(link))
			}
		}
	})
	// 查看文章内容
	c.OnHTML("td#UCAP-CONTENT", func(e *colly.HTMLElement) {
		s := strings.Split(e.Text, "\n")
		// 去掉前面无效的信息
		arr := make([]string, 0)
		for i, v := range s {
			if strings.Contains(v, "一、") {
				arr = s[i:]
				break
			}
		}
		for _, v1 := range arr {
			for _, v2 := range legalHolidays {
				if strings.Contains(v1, v2) {
					lineArr := strings.Split(v1, "。")
					// 截取休假时间
					// 1月1日至3日
					// 2月11日至17日
					item0 := lineArr[0]
					r1 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日至([\d]{1,2})月([\d]{1,2})日`)
					r2 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日至([\d]{1,2})日`)
					var holidayDates []string
					if r1.MatchString(item0) {
						params := r1.FindStringSubmatch(item0)
						if len(params) == 5 {
							startMonth := str2Int(params[1])
							endMonth := str2Int(params[3])
							holidayDates = getDates(
								fmt.Sprintf("%d-%02d-%02d", year, startMonth, str2Int(params[2])),
								fmt.Sprintf("%d-%02d-%02d", year, endMonth, str2Int(params[4])),
							)
						}
					} else if r2.MatchString(item0) {
						params := r2.FindStringSubmatch(item0)
						if len(params) == 4 {
							month := str2Int(params[1])
							holidayDates = getDates(
								fmt.Sprintf("%d-%02d-%02d", year, month, str2Int(params[2])),
								fmt.Sprintf("%d-%02d-%02d", year, month, str2Int(params[3])),
							)
						}
					}
					// 去重
					for _, date := range holidayDates {
						if !funk.ContainsString(holidays, date) {
							holidays = append(holidays, date)
						}
					}

					// 截取调休时间
					if len(lineArr) > 2 {
						item1 := lineArr[1]
						item1Arr := strings.Split(item1, "、")
						for _, v3 := range item1Arr {
							// r3解释: GOV一般是11月份发布第二年, 可能涉及跨年调休, 该调休应该属于上一年
							r3 := regexp.MustCompile(`([\d]{4})年([\d]{1,2})月([\d]{1,2})日`)
							r4 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日`)
							if r3.MatchString(v3) {
								params := r3.FindStringSubmatch(v3)
								if len(params) == 4 {
									if str2Int(params[1]) == year-1 {
										date := fmt.Sprintf("%d-%02d-%02d", year-1, str2Int(params[2]), str2Int(params[3]))
										if !funk.ContainsString(lastYearWorkdays, date) {
											lastYearWorkdays = append(lastYearWorkdays, date)
										}
									}
								}
							} else if r4.MatchString(v3) {
								params := r4.FindStringSubmatch(v3)
								if len(params) == 3 {
									date := fmt.Sprintf("%d-%02d-%02d", year, str2Int(params[1]), str2Int(params[2]))
									if !funk.ContainsString(workdays, date) {
										workdays = append(workdays, date)
									}
								}
							}
						}
					}
				}
			}
		}
	})

	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	// 访问完成
	c.OnScraped(func(r *colly.Response) {
		if r.Request.URL.String() == searchUrl {
			wg.Done()
		}
	})
	c.OnError(func(r *colly.Response, e error) {
		if r.Request.URL.String() == searchUrl {
			err = e
			wg.Done()
		}
	})

	// 访问页面
	c.Visit(searchUrl)

	// 等待结束
	c.Wait()
	if len(lastYearWorkdays) > 0 {
		oldLastYearHolidays, oldLastYearWorkdays, lastYearErr := s.all(year - 1)
		if lastYearErr == nil && s.f {
			oldLastYearWorkdays = append(oldLastYearWorkdays, lastYearWorkdays...)
			s.refreshFile(year-1, oldLastYearHolidays, oldLastYearWorkdays)
		}
	}
	return holidays, workdays, err
}

// 获取两日期之间的全部年份
func getYears(start, end string) []int {
	res := make([]int, 0)
	startTime, endTime := getTimeRanges(start, end)
	if startTime == nil || endTime == nil {
		return res
	}
	endYear := endTime.Year()
	res = append(res, startTime.Year())
	// 刚好只有一年
	if endYear == startTime.Year() {
		return res
	}
	for {
		current := startTime.AddDate(1, 0, 0)
		dateYear := current.Year()
		startTime = &current
		res = append(res, dateYear)
		if dateYear == endYear {
			break
		}
	}
	return res
}

// 获取两日期之间的全部日期
func getDates(start, end string) []string {
	res := make([]string, 0)
	startTime, endTime := getTimeRanges(start, end)
	if startTime == nil || endTime == nil {
		return res
	}
	// 输出日期格式固定
	timeFormatTpl := dateFormat
	endStr := endTime.Format(timeFormatTpl)
	res = append(res, startTime.Format(timeFormatTpl))
	for {
		current := startTime.AddDate(0, 0, 1)
		dateStr := startTime.Format(timeFormatTpl)
		startTime = &current
		res = append(res, dateStr)
		if dateStr == endStr {
			break
		}
	}
	return res
}

// 时间开始结束范围
func getTimeRanges(start, end string) (*time.Time, *time.Time) {
	startTime, endTime, after := timeAfter(start, end)
	if startTime != nil && endTime != nil && !after {
		t := startTime
		startTime = endTime
		endTime = t
		return startTime, endTime
	}
	return startTime, endTime
}

// 两日期字符串比较
func timeAfter(start, end string) (*time.Time, *time.Time, bool) {
	timeFormatTpl := timeFormat
	if len(timeFormatTpl) != len(start) {
		timeFormatTpl = timeFormatTpl[0:len(start)]
	}
	startTime, err := time.Parse(timeFormatTpl, start)
	if err != nil {
		return nil, nil, false
	}
	endTime, err := time.Parse(timeFormatTpl, end)
	if err != nil {
		return nil, nil, false
	}
	if endTime.After(startTime) {
		return &startTime, &endTime, true
	}
	return &startTime, &endTime, false
}

// 字符串转int
func str2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}
