package main

import (
	"testing"
	"time"
)

func TestPermalinkTitle(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-11")
	url := getPostUrl("my-blog", date, []string{}, "/:title/")

	if (url != "/my-blog/") {
		t.Errorf("Expected url [/my-blog/] got [%s]", url)
	}
}

func TestPermalinkYear(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-11")
	url := getPostUrl("my-blog", date, []string{}, "/:year/")

	if (url != "/2013/") {
		t.Errorf("Expected url [/2013/] got [%s]", url)
	}
}

func TestPermalinkMonth(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-11")
	url := getPostUrl("my-blog", date, []string{}, "/:month/")

	if (url != "/11/") {
		t.Errorf("Expected url [/11/] got [%s]", url)
	}
}

func TestPermalinkMonthWithOneDigitMonth(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-02-11")
	url := getPostUrl("my-blog", date, []string{}, "/:month/")

	if (url != "/02/") {
		t.Errorf("Expected url [/02/] got [%s]", url)
	}
}

func TestPermalinkIMonth(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	url := getPostUrl("my-blog", date, []string{}, "/:i_month/")

	if (url != "/11/") {
		t.Errorf("Expected url [/11/] got [%s]", url)
	}
}

func TestPermalinkIMonthWithOneDigit(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-02-01")
	url := getPostUrl("my-blog", date, []string{}, "/:i_month/")

	if (url != "/2/") {
		t.Errorf("Expected url [/2/] got [%s]", url)
	}
}

func TestPermalinkDay(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-25")
	url := getPostUrl("my-blog", date, []string{}, "/:day/")

	if (url != "/25/") {
		t.Errorf("Expected url [/25/] got [%s]", url)
	}
}

func TestPermalinkDayWithOneDigit(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	url := getPostUrl("my-blog", date, []string{}, "/:day/")

	if (url != "/01/") {
		t.Errorf("Expected url [/01/] got [%s]", url)
	}
}


func TestPermalinkIDay(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-25")
	url := getPostUrl("my-blog", date, []string{}, "/:i_day/")

	if (url != "/25/") {
		t.Errorf("Expected url [/25/] got [%s]", url)
	}
}

func TestPermalinkIDayWithOneDigit(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	url := getPostUrl("my-blog", date, []string{}, "/:i_day/")

	if (url != "/1/") {
		t.Errorf("Expected url [/1/] got [%s]", url)
	}
}

func TestSingleCategory(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{"My Category"}
	url := getPostUrl("my-blog", date, categories, "/:categories/index.html")

	if (url != "/my-category/index.html") {
		t.Errorf("Expected url [/my-category/index.html] got [%s]", url)
	}
}

func TestMultipleCategories(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{"Go Lang", "Unit Testing"}
	url := getPostUrl("my-blog", date, categories, "/:categories/index.html")

	if (url != "/go-lang/unit-testing/index.html") {
		t.Errorf("Expected url [/go-lang/unit-testing/index.html] got [%s]", url)
	}
}

func TestEmptyCategories(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{}
	url := getPostUrl("my-blog", date, categories, "/:categories/:title/")

	if (url != "/my-blog/") {
		t.Errorf("Expected url [/my-blog/] got [%s]", url)
	}
}

func TestUnicodeCategories(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{"áãàâäçéèëíìïóòôöúùûü"}
	url := getPostUrl("my-blog", date, categories, "/:categories/:title/")

	if (url != "/aaaaaceeeiiioooouuuu/my-blog/") {
		t.Errorf("Expected url [/my-blog/] got [%s]", url)
	}
}


// The 'date' permalink is a shortcut to /:categories/:year/:month/:day/:title.html
func TestDateBuiltinPermakink(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{"go"}
	url := getPostUrl("unit-tests", date, categories, "date")

	if (url != "/go/2013/11/01/unit-tests.html") {
		t.Errorf("Expected url [/go/2013/11/01/unit-tests.html] got [%s]", url)
	}
}

// The 'pretty' permalink is a shortcut to /:categories/:year/:month/:day/:title/
func TestPrettyBuiltinPermakink(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{"go"}
	url := getPostUrl("unit-tests", date, categories, "pretty")

	if (url != "/go/2013/11/01/unit-tests/") {
		t.Errorf("Expected url [/go/2013/11/01/unit-tests/] got [%s]", url)
	}
}

// The 'none' permalink is a shortcut to /:categories/:title.html
func TestNoneBuiltinPermakink(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2013-11-01")
	categories := []string{"go"}
	url := getPostUrl("unit-tests", date, categories, "none")

	if (url != "/go/unit-tests.html") {
		t.Errorf("Expected url [/go/unit-tests.html] got [%s]", url)
	}
}
