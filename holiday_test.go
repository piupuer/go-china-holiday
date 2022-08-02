package holiday

import (
	"testing"
)

func TestChinaHoliday_Check(t *testing.T) {
	type fields struct {
		ops Options
		f   bool
	}
	type args struct {
		date string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantHoliday bool
	}{
		{
			name:   "test0",
			fields: fields{},
			args: args{
				date: "2018-01-01",
			},
			wantHoliday: true,
		},
		{
			name:   "test1",
			fields: fields{},
			args: args{
				date: "2019-01-01",
			},
			wantHoliday: true,
		},
		{
			name:   "test2",
			fields: fields{},
			args: args{
				date: "2020-01-01",
			},
			wantHoliday: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := ChinaHoliday{
				ops: tt.fields.ops,
				f:   tt.fields.f,
			}
			if gotHoliday := ch.Check(tt.args.date); gotHoliday != tt.wantHoliday {
				t.Errorf("Check() = %v, want %v", gotHoliday, tt.wantHoliday)
			}
		})
	}
}
