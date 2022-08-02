package holiday

const (
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

const (
	I1 = iota
	I2
	I3
	I4
	I5
	I6
	I7
	I8
	I9
	I10
	I11
	I12
	I13
	I14
	I15
)

var IndexMap = map[int]string{
	I1:  "一",
	I2:  "二",
	I3:  "三",
	I4:  "四",
	I5:  "五",
	I6:  "六",
	I7:  "七",
	I8:  "八",
	I9:  "九",
	I10: "十",
	I11: "十一",
	I12: "十二",
	I13: "十三",
	I14: "十四",
	I15: "十五",
}
