package model

import (
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestReport_BeforeSave(t *testing.T) {
	var reports = []struct {
		name        string
		in          *Report
		expected    *Report
		expectError bool
	}{
		{"no project id", &Report{}, &Report{}, true},
		{"with project id", &Report{ProjectID: 123}, &Report{ProjectID: 123, Status: "ok"}, false},
	}

	for _, tt := range reports {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.BeforeSave(nil)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, tt.in)
		})
	}
}

func TestReport(t *testing.T) {
	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	db.Exec("PRAGMA foreign_keys = ON;")
	db.AutoMigrate(&Project{}, &Report{}, &Detail{})

	var rid, pid uint

	t.Run("create", func(t *testing.T) {
		p := Project{
			Name:        "Test Project 111 ",
			Description: "Test Description Asdf ",
		}

		r := Report{
			Project:   &p,
			Name:      "Test report",
			EndReason: "normal",
			Date:      time.Now(),
			Count:     200,
			Total:     time.Duration(2 * time.Second),
			Average:   time.Duration(10 * time.Millisecond),
			Fastest:   time.Duration(1 * time.Millisecond),
			Slowest:   time.Duration(100 * time.Millisecond),
			Rps:       2000,
		}

		// r.Options = &Options{
		// 	Name:        "Test report",
		// 	Call:        "helloworld.Greeter.SayHello",
		// 	Proto:       "../../testdata/greeter.proto",
		// 	Host:        "0.0.0.0:50051",
		// 	N:           200,
		// 	C:           50,
		// 	Timeout:     time.Duration(20 * time.Second),
		// 	DialTimeout: time.Duration(10 * time.Second),
		// 	CPUs:        8,
		// 	Insecure:    true,
		// 	Data:        map[string]string{"name": "Joe"},
		// 	Metadata:    &map[string]string{"token": "abc123", "request-id": "12345"},
		// }

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            3,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         3,
			"DeadlineExceeded": 2}

		r.Tags = map[string]string{
			"env":        "staging",
			"created by": "Joe Developer",
		}

		err := db.Create(&r).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		pid = p.ID
		rid = r.ID

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotNil(t, p2.CreatedAt)
		assert.NotNil(t, p2.UpdatedAt)
		assert.Nil(t, p2.DeletedAt)
	})

	t.Run("read", func(t *testing.T) {
		r := new(Report)
		err = db.First(r, rid).Error

		assert.NoError(t, err)

		assert.Equal(t, pid, r.ProjectID)
		assert.NotNil(t, r.CreatedAt)
		assert.NotNil(t, r.UpdatedAt)
		assert.Nil(t, r.DeletedAt)
		assert.Equal(t, StatusOK, r.Status)

		assert.Equal(t, "Test report", r.Name)
		assert.Equal(t, "normal", r.EndReason)
		assert.NotZero(t, r.Date)

		assert.Equal(t, 3, r.ErrorDist["rpc error: code = Internal desc = Internal error."])
		assert.Equal(t, 2, r.ErrorDist["rpc error: code = DeadlineExceeded desc = Deadline exceeded."])

		assert.Equal(t, 195, r.StatusCodeDist["OK"])
		assert.Equal(t, 3, r.StatusCodeDist["Internal"])
		assert.Equal(t, 2, r.StatusCodeDist["DeadlineExceeded"])

		// assert.Equal(t, "Test report", r.Options.Name)
		// assert.Equal(t, "helloworld.Greeter.SayHello", r.Options.Call)
		// assert.Equal(t, "../../testdata/greeter.proto", r.Options.Proto)
		// assert.Equal(t, "0.0.0.0:50051", r.Options.Host)
		// assert.Equal(t, uint(200), r.Options.N)
		// assert.Equal(t, uint(50), r.Options.C)
		// assert.Equal(t, time.Duration(20*time.Second), r.Options.Timeout)
		// assert.Equal(t, time.Duration(10*time.Second), r.Options.DialTimeout)
		// assert.Equal(t, map[string]interface{}{"name": "Joe"}, r.Options.Data)
		// assert.Equal(t, &map[string]string{"token": "abc123", "request-id": "12345"}, r.Options.Metadata)
		// assert.Equal(t, false, r.Options.Binary)
		// assert.Equal(t, true, r.Options.Insecure)
		// assert.Equal(t, 8, r.Options.CPUs)

		assert.Equal(t, "staging", r.Tags["env"])
		assert.Equal(t, "Joe Developer", r.Tags["created by"])
	})

	t.Run("create with project id", func(t *testing.T) {
		r := Report{
			ProjectID: pid,
			Name:      "Test report 2",
			EndReason: "cancelled",
			Date:      time.Now(),
			Count:     300,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(11 * time.Millisecond),
			Fastest:   time.Duration(2 * time.Millisecond),
			Slowest:   time.Duration(120 * time.Millisecond),
			Rps:       2100,
		}

		err := db.Create(&r).Error

		assert.NoError(t, err)
		assert.NotZero(t, r.ID)

		r2 := new(Report)
		err = db.First(r2, r.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, r.Name, r2.Name)
		assert.Equal(t, r.EndReason, r2.EndReason)
		assert.Equal(t, StatusOK, r2.Status)
		assert.NotNil(t, r2.CreatedAt)
		assert.NotNil(t, r2.UpdatedAt)
		assert.Nil(t, r2.DeletedAt)
		assert.Equal(t, uint64(300), r2.Count)
		assert.Equal(t, float64(2100), r2.Rps)
	})

	t.Run("fail with invalid project id", func(t *testing.T) {
		r := Report{
			ProjectID: 123432,
			Name:      "Test report 2",
			EndReason: "cancelled",
			Date:      time.Now(),
			Count:     300,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(11 * time.Millisecond),
			Fastest:   time.Duration(2 * time.Millisecond),
			Slowest:   time.Duration(120 * time.Millisecond),
			Rps:       2100,
		}

		err := db.Create(&r).Error

		assert.Error(t, err)
	})
}