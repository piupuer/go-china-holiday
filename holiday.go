package holiday

import (
	"bufio"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/golang-module/carbon/v2"
	log "github.com/sirupsen/logrus"
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

type ChinaHoliday struct {
	ops Options
	f   bool
}

func New(options ...func(*Options)) (ins ChinaHoliday) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.filename != "" {
		ins.f = true
		info, err := os.Stat(ops.filename)
		var file *os.File
		if err != nil {
			// create file if not exists
			file, err = os.OpenFile(ops.filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				ins.f = false
			}
		} else if info.IsDir() {
			// skip dir
			ins.f = false
		}
		defer file.Close()
	}
	ins.ops = *ops
	return
}

// Check 检查某个日期是否是节假日(支持yyyy-MM-dd)
func (ch ChinaHoliday) Check(date string) (holiday bool) {
	d := carbon.Parse(date)

	year := d.Year()
	holidays, workdays := ch.all(year)

	// 校验是否节假日
	nowDate := d.ToDateString()
	if funk.ContainsString(holidays, nowDate) {
		// 节假日
		holiday = true
		return
	} else {
		if d.IsWeekday() && !funk.ContainsString(workdays, nowDate) {
			// 周末且不是调休
			holiday = true
			return
		}
	}
	return
}

// List 获取某一年节假日数目(周末除外)
func (ch ChinaHoliday) List(years ...int) (holidays, workdays []string) {
	for _, year := range years {
		h, w := ch.all(year)
		holidays = append(holidays, h...)
		workdays = append(workdays, w...)
	}
	return
}

// Range 获取指定日期之间有多少节假日
func (ch ChinaHoliday) Range(start, end string) (holidays, workdays []string) {
	years := getYears(start, end)
	allHolidays, allWorkdays := ch.all(years...)
	holidays = make([]string, 0)
	workdays = make([]string, 0)
	for _, holiday := range allHolidays {
		if lt(start, holiday) && lt(holiday, end) {
			holidays = append(holidays, holiday)
		}
	}
	for _, workday := range allWorkdays {
		if lt(start, workday) && lt(workday, end) {
			workdays = append(workdays, workday)
		}
	}
	return
}

// 读取全部节假日(一年或多年)
func (ch ChinaHoliday) all(years ...int) (holidays, workdays []string) {
	holidays = make([]string, 0)
	workdays = make([]string, 0)
	// 从文件中读取数据
	list := ch.getFromFile()
	for _, year := range years {
		currentHolidays := make([]string, 0)
		currentWorkdays := make([]string, 0)
		if oldData, ok := list[year]; ok {
			if len(oldData) == 1 {
				currentHolidays = oldData[0]
			} else if len(oldData) == 2 {
				currentHolidays = oldData[0]
				currentWorkdays = oldData[1]
			}
		}
		if len(currentHolidays) == 0 && len(currentWorkdays) == 0 {
			// 读取线上数据
			currentHolidays, currentWorkdays = ch.online(year)
			if ch.f {
				// 存入文件
				ch.appendToFile(year, currentHolidays, currentWorkdays)
			}
		}
		if len(currentHolidays) > 0 {
			holidays = append(holidays, currentHolidays...)
		}
		if len(currentWorkdays) > 0 {
			workdays = append(workdays, currentWorkdays...)
		}
	}
	return
}

// 从文件中获取数据
func (ch ChinaHoliday) getFromFile() (list map[int][][]string) {
	list = make(map[int][][]string)
	if !ch.f {
		return
	}
	file, err := os.OpenFile(ch.ops.filename, os.O_RDONLY, 0666)
	if err != nil {
		return
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
	return
}

func (ch ChinaHoliday) appendToFile(year int, holidays, workdays []string) {
	file, err := os.OpenFile(ch.ops.filename, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	write := bufio.NewWriter(file)
	write.WriteString(fmt.Sprintf("%d\n", year))
	write.WriteString(fmt.Sprintf("%s\n", strings.Join(holidays, ",")))
	write.WriteString(fmt.Sprintf("%s\n", strings.Join(workdays, ",")))
	write.WriteString(fmt.Sprintf("%s\n", fileSp))
	write.Flush()
	return
}

func (ch ChinaHoliday) refreshFile(year int, holidays, workdays []string) error {
	b, err := ioutil.ReadFile(ch.ops.filename)
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
	file, err := os.OpenFile(ch.ops.filename, os.O_WRONLY|os.O_TRUNC, 0666)
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
func (ch ChinaHoliday) online(year int) (holidays, workdays []string) {
	time.Sleep(time.Second)
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
	holidays = make([]string, 0)
	// 调休时间
	workdays = make([]string, 0)
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
		arr := convertText(e.Text)
		for _, v1 := range arr {
			for _, v2 := range legalHolidays {
				if strings.Contains(v1, v2) {
					lineArr := strings.Split(v1, "。")
					// 截取休假时间
					// 1月1日至3日
					// 2月11日至17日
					// 1月1日
					item0 := lineArr[0]
					r1 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日至([\d]{1,2})月([\d]{1,2})日`)
					r2 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日至([\d]{1,2})日`)
					r3 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日`)
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
					} else if r3.MatchString(item0) {
						params := r3.FindStringSubmatch(item0)
						if len(params) == 3 {
							month := str2Int(params[1])
							holidayDates = append(holidayDates, fmt.Sprintf("%d-%02d-%02d", year, month, str2Int(params[2])))
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
							// r4解释: GOV一般是11月份发布第二年, 可能涉及跨年调休, 该调休应该属于上一年
							r4 := regexp.MustCompile(`([\d]{4})年([\d]{1,2})月([\d]{1,2})日`)
							r5 := regexp.MustCompile(`([\d]{1,2})月([\d]{1,2})日`)
							if r4.MatchString(v3) {
								params := r4.FindStringSubmatch(v3)
								if len(params) == 4 {
									if str2Int(params[1]) == year-1 {
										date := fmt.Sprintf("%d-%02d-%02d", year-1, str2Int(params[2]), str2Int(params[3]))
										if !funk.ContainsString(lastYearWorkdays, date) {
											lastYearWorkdays = append(lastYearWorkdays, date)
										}
									}
								}
							} else if r5.MatchString(v3) {
								params := r5.FindStringSubmatch(v3)
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

	c.Wait()
	if len(lastYearWorkdays) > 0 {
		oldLastYearHolidays, oldLastYearWorkdays := ch.all(year - 1)
		if ch.f {
			oldLastYearWorkdays = append(oldLastYearWorkdays, lastYearWorkdays...)
			ch.refreshFile(year-1, oldLastYearHolidays, oldLastYearWorkdays)
		}
	}
	if err != nil {
		log.Warn(err)
	}
	return
}

func convertText(text string) (arr []string) {
	l := len(IndexMap)
	for i := 0; i < l; i++ {
		item := IndexMap[i]
		i1 := strings.Index(text, fmt.Sprintf("%s、", item))
		if i1 >= 0 {
			s1 := text[i1:]
			if i < l-1 {
				nextItem := IndexMap[i+1]
				i2 := strings.Index(s1, fmt.Sprintf("%s、", nextItem))
				in := strings.Index(s1, "\n")
				if i2 >= 0 {
					arr = append(arr, s1[0:i2])
				} else if in >= 0 {
					arr = append(arr, s1[0:in])
				} else {
					arr = append(arr, s1)
				}
			}
		}
	}
	return
}

// 获取两日期之间的全部年份
func getYears(start, end string) (years []int) {
	years = make([]int, 0)
	startTime, endTime := carbon.Parse(start), carbon.Parse(end)
	if startTime.IsInvalid() || endTime.IsInvalid() {
		return
	}
	if startTime.Gte(endTime) {
		return
	}
	endYear := endTime.Year()
	years = append(years, startTime.Year())
	if endYear == startTime.Year() {
		return
	}
	current := startTime.AddYear()
	for {
		currentYear := current.Year()
		years = append(years, currentYear)
		if currentYear == endYear {
			break
		}
		current = current.AddYear()
	}
	return
}

// 获取两日期之间的全部日期
func getDates(start, end string) (dates []string) {
	dates = make([]string, 0)
	startTime, endTime := carbon.Parse(start), carbon.Parse(end)
	if startTime.IsInvalid() || endTime.IsInvalid() {
		return
	}
	if startTime.Gte(endTime) {
		return
	}
	endDate := endTime.ToDateString()
	dates = append(dates, startTime.ToDateString())
	if endDate == startTime.ToDateString() {
		return
	}
	current := startTime.AddDay()
	for {
		currentDate := current.ToDateString()
		dates = append(dates, currentDate)
		if currentDate == endDate {
			break
		}
		current = current.AddDay()
	}
	return
}

// 两日期字符串比较
func lt(start, end string) (ok bool) {
	startTime, endTime := carbon.Parse(start), carbon.Parse(end)
	if startTime.IsInvalid() || endTime.IsInvalid() {
		return
	}
	ok = startTime.Lt(endTime)
	return
}

func str2Int(str string) (num int) {
	num, _ = strconv.Atoi(str)
	return
}
