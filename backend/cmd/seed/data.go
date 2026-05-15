package main

type CountryConfig struct {
	Name     string
	Code     string
	Currency string
	MinBase  float64
	MaxBase  float64
	Weight   int // relative probability of an employee being in this country
}

// Countries used by both the seeder and reference data setup.
var Countries = []CountryConfig{
	{"United States", "US", "USD", 60000, 200000, 25},
	{"India", "IN", "INR", 400000, 4000000, 30},
	{"United Kingdom", "GB", "GBP", 35000, 120000, 10},
	{"Germany", "DE", "EUR", 45000, 130000, 10},
	{"Canada", "CA", "CAD", 55000, 160000, 7},
	{"Australia", "AU", "AUD", 65000, 170000, 5},
	{"France", "FR", "EUR", 40000, 110000, 5},
	{"Japan", "JP", "JPY", 4000000, 15000000, 5},
	{"Singapore", "SG", "SGD", 50000, 180000, 3},
}

// Departments to seed.
var Departments = []string{
	"Engineering", "Product", "Design", "Data Science",
	"Marketing", "Sales", "Human Resources", "Finance",
	"Operations", "Customer Success", "Legal", "IT Infrastructure",
}

// JobTitlesByDepartment maps a department name to the job titles that belong
// to it. The seeder uses this to create the job_titles rows linked by FK.
var JobTitlesByDepartment = map[string][]string{
	"Engineering":       {"Software Engineer", "Senior Software Engineer", "Staff Engineer", "Engineering Manager", "Director of Engineering"},
	"Product":           {"Product Manager", "Senior Product Manager", "Product Lead"},
	"Design":            {"UX Designer", "Senior UX Designer", "Design Lead"},
	"Data Science":      {"Data Analyst", "Senior Data Analyst", "Data Scientist"},
	"Marketing":         {"Marketing Specialist", "Marketing Manager", "Content Strategist"},
	"Sales":             {"Sales Representative", "Account Executive", "Sales Manager"},
	"Human Resources":   {"HR Specialist", "HR Manager"},
	"Finance":           {"Finance Analyst", "Financial Controller"},
	"Operations":        {"Operations Analyst", "Project Manager", "Business Analyst"},
	"Customer Success":  {"Customer Success Manager", "Support Engineer"},
	"Legal":             {"Legal Counsel", "Compliance Officer"},
	"IT Infrastructure": {"DevOps Engineer", "Senior DevOps Engineer", "Site Reliability Engineer"},
}

var FirstNames = []string{
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

var LastNames = []string{
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

var Streets = []string{
	"Main St", "Oak Ave", "Maple Dr", "Cedar Ln", "Park Blvd",
	"Elm St", "Pine Rd", "Lake View", "Hill Top", "River Rd",
}
