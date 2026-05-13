package seed

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

var firstNames = []string{
	"James", "Mary", "Robert", "Patricia", "John", "Jennifer", "Michael", "Linda",
	"David", "Elizabeth", "William", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
	"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Lisa", "Daniel", "Nancy",
	"Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra", "Donald", "Ashley",
	"Steven", "Kimberly", "Paul", "Emily", "Andrew", "Donna", "Joshua", "Michelle",
	"Amit", "Priya", "Rahul", "Sunita", "Vikram", "Anita", "Rajesh", "Deepa",
	"Sanjay", "Kavita", "Wei", "Mei", "Jun", "Yan", "Chen", "Li",
	"Hans", "Anna", "Klaus", "Petra", "Friedrich", "Helga", "Otto", "Ingrid",
	"Pierre", "Marie", "Jean", "Sophie", "Louis", "Camille", "Henri", "Isabelle",
	"Takeshi", "Yuki", "Kenji", "Haruka", "Satoshi", "Sakura", "Hiroshi", "Aiko",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
	"Rodriguez", "Martinez", "Hernandez", "Lopez", "Wilson", "Anderson", "Thomas",
	"Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
	"Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker",
	"Sharma", "Patel", "Kumar", "Singh", "Gupta", "Verma", "Mehta", "Joshi",
	"Wang", "Zhang", "Li", "Chen", "Liu", "Yang", "Huang", "Wu",
	"Mueller", "Schmidt", "Fischer", "Weber", "Meyer", "Wagner", "Becker", "Schulz",
	"Dubois", "Laurent", "Bernard", "Moreau", "Petit", "Roux", "Leroy", "David",
	"Tanaka", "Suzuki", "Watanabe", "Yamamoto", "Nakamura", "Kobayashi", "Saito", "Kato",
}

var jobTitles = []string{
	"Software Engineer", "Senior Software Engineer", "Staff Engineer",
	"Product Manager", "Senior Product Manager",
	"Data Analyst", "Senior Data Analyst", "Data Scientist",
	"UX Designer", "Senior UX Designer",
	"DevOps Engineer", "Senior DevOps Engineer",
	"QA Engineer", "Senior QA Engineer",
	"Engineering Manager", "Director of Engineering",
	"Marketing Manager", "Marketing Specialist",
	"Sales Representative", "Account Executive",
	"HR Specialist", "HR Manager",
	"Finance Analyst", "Financial Controller",
	"Technical Writer", "Content Strategist",
	"Customer Success Manager", "Support Engineer",
	"Business Analyst", "Project Manager",
}

var departments = []string{
	"Engineering", "Product", "Design", "Data Science",
	"Marketing", "Sales", "Human Resources", "Finance",
	"Operations", "Customer Success", "Legal", "IT Infrastructure",
}

type countryConfig struct {
	Name     string
	Currency string
	MinBase  float64
	MaxBase  float64
	Weight   int // relative probability of employees in this country
}

var countries = []countryConfig{
	{"United States", "USD", 60000, 200000, 25},
	{"India", "INR", 400000, 4000000, 30},
	{"United Kingdom", "GBP", 35000, 120000, 10},
	{"Germany", "EUR", 45000, 130000, 10},
	{"Canada", "CAD", 55000, 160000, 7},
	{"Australia", "AUD", 65000, 170000, 5},
	{"France", "EUR", 40000, 110000, 5},
	{"Japan", "JPY", 4000000, 15000000, 5},
	{"Singapore", "SGD", 50000, 180000, 3},
}

func GenerateEmployees(repo repository.EmployeeRepository, count int) error {
	ctx := context.Background()
	rng := rand.New(rand.NewSource(42)) // deterministic seed for reproducibility

	totalWeight := 0
	for _, c := range countries {
		totalWeight += c.Weight
	}

	emailSet := make(map[string]bool)

	for i := 0; i < count; i++ {
		firstName := firstNames[rng.Intn(len(firstNames))]
		lastName := lastNames[rng.Intn(len(lastNames))]
		fullName := firstName + " " + lastName

		email := generateUniqueEmail(firstName, lastName, i, emailSet)

		country := pickCountry(rng, totalWeight)
		salary := country.MinBase + rng.Float64()*(country.MaxBase-country.MinBase)
		salary = float64(int(salary/100)) * 100 // round to nearest 100

		joinYear := 2015 + rng.Intn(10)
		joinMonth := time.Month(1 + rng.Intn(12))
		joinDay := 1 + rng.Intn(28)

		emp := &model.Employee{
			FullName:   fullName,
			Email:      email,
			JobTitle:   jobTitles[rng.Intn(len(jobTitles))],
			Department: departments[rng.Intn(len(departments))],
			Country:    country.Name,
			Salary:     salary,
			Currency:   country.Currency,
			JoinDate:   time.Date(joinYear, joinMonth, joinDay, 0, 0, 0, 0, time.UTC),
		}

		if err := repo.Create(ctx, emp); err != nil {
			return fmt.Errorf("create employee %d: %w", i, err)
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("  Seeded %d/%d employees\n", i+1, count)
		}
	}

	return nil
}

func generateUniqueEmail(first, last string, index int, seen map[string]bool) string {
	base := fmt.Sprintf("%s.%s", toLower(first), toLower(last))
	email := base + "@company.com"
	if !seen[email] {
		seen[email] = true
		return email
	}
	email = fmt.Sprintf("%s.%s%d@company.com", toLower(first), toLower(last), index)
	seen[email] = true
	return email
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func pickCountry(rng *rand.Rand, totalWeight int) countryConfig {
	r := rng.Intn(totalWeight)
	for _, c := range countries {
		r -= c.Weight
		if r < 0 {
			return c
		}
	}
	return countries[0]
}
