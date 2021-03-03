package holiday

import (
	"fmt"
	"strings"
	"testing"
)

func TestChinaHoliday(t *testing.T) {
	ins, err := New(&Config{})
	if err != nil {
		panic(err)
	}
	b, err := ins.Check("2021-03-02")
	if err != nil {
		panic(err)
	}
	fmt.Println("2021-03-02是否节假日:", b)

	b2, err := ins.Check("2020-10-01")
	if err != nil {
		panic(err)
	}
	fmt.Println("2020-10-01是否节假日:", b2)

	h1, w1, err := ins.List(2012)
	if err != nil {
		panic(err)
	}
	fmt.Println("2012全年节假日:", strings.Join(h1, ","), "调休日:", strings.Join(w1, ","))

	h2, w2, err := ins.Range("2021-03-01", "2021-05-01")
	if err != nil {
		panic(err)
	}
	fmt.Println("2021全年节假日2021-03-01到2021-05-01:", strings.Join(h2, ","), "调休日:", strings.Join(w2, ","))
}
