package scraper

import (
	"strings"
	"testing"
)

const fixture = `
<div class="js-vacancies-sections">
  <section class="vacancies-section">
    <div class="vacancies-section__items">
      <div class="vacancies-section__item"
           data-vacancy-id="7509"
           data-vacancy-section="Разработка"
           data-vacancy-geo="Москва"
           data-vacancy-team="Техническая платформа"
           data-vacancy-remote="Да">
        <a href="/vacancies/razrabotka/19456/" class="vacancies-section__item-link"></a>
        <div class="vacancies-section__item-content">
          <a href="/vacancies/razrabotka/19456/" class="vacancies-section__item-name">Android-разработчик</a>
        </div>
        <div class="vacancies-section__item-cities">Москва</div>
      </div>
      <div class="vacancies-section__item"
           data-vacancy-id="7510"
           data-vacancy-geo="Санкт-Петербург"
           data-vacancy-team="Backend">
        <a href="/vacancies/razrabotka/19457/" class="vacancies-section__item-name">Go-разработчик</a>
      </div>
    </div>
  </section>
</div>`

func TestParse(t *testing.T) {
	vs, err := Parse(strings.NewReader(fixture), "https://career.avito.com/vacancies/")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 2 {
		t.Fatalf("expected 2 vacancies, got %d", len(vs))
	}
	v := vs[0]
	if v.ID != "7509" || v.Name != "Android-разработчик" {
		t.Errorf("unexpected vacancy: %+v", v)
	}
	if v.URL != "https://career.avito.com/vacancies/razrabotka/19456/" {
		t.Errorf("unexpected url: %s", v.URL)
	}
	if v.City != "Москва" || v.Team != "Техническая платформа" || v.Remote != "Да" {
		t.Errorf("unexpected meta: %+v", v)
	}
}

func TestMatches(t *testing.T) {
	vs, err := Parse(strings.NewReader(fixture), "https://career.avito.com/vacancies/")
	if err != nil {
		t.Fatal(err)
	}
	kw := []string{"android"}
	if !vs[0].Matches(kw) {
		t.Error("expected android vacancy to match")
	}
	if vs[1].Matches(kw) {
		t.Error("expected go vacancy not to match")
	}
	if !vs[0].Matches([]string{"ANDROID"}) {
		t.Error("matching should be case-insensitive")
	}
}
