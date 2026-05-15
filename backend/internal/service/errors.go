package service

import "errors"

// Sentinel errors returned by the service layer.
//
// They fall into two groups:
//
//   - Result errors (ErrNotFound, ErrDuplicateEmail, ErrInvalidInput) are
//     returned to indicate the outcome of an operation. Handlers map each to
//     a specific HTTP status (404, 409, 400 respectively).
//
//   - Validation errors (the Err*Required, Err*Invalid family) are returned
//     when the input fails a business-rule check. They all map to HTTP 400.
//     The handler uses IsValidationError to recognise the whole family in
//     one call instead of duplicating the message strings.
//
// Defining these as exported variables (rather than ad-hoc errors.New calls
// at each return site) keeps the public error contract in one place and
// lets callers compare with errors.Is.
var (
	// Result errors.
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotFound       = errors.New("employee not found")
	ErrDuplicateEmail = errors.New("email already exists")

	// Employee field validation.
	ErrFirstNameRequired = errors.New("first name is required")
	ErrLastNameRequired  = errors.New("last name is required")
	ErrEmailRequired     = errors.New("email is required")
	ErrEmailInvalid      = errors.New("email must be valid")
	ErrJobTitleRequired  = errors.New("job title is required")
	ErrCountryRequired   = errors.New("country is required")
	ErrSalaryNegative    = errors.New("salary must be non-negative")

	// Foreign-key validation (referenced row missing or inactive).
	ErrCountryInactive    = errors.New("country does not exist or is inactive")
	ErrJobTitleInactive   = errors.New("job title does not exist or is inactive")
	ErrDepartmentInactive = errors.New("department does not exist or is inactive")

	// Reference-data field validation.
	ErrCountryNameRequired   = errors.New("country name is required")
	ErrCountryCodeRequired   = errors.New("country code is required")
	ErrCurrencyRequired      = errors.New("currency is required")
	ErrDepartmentNameRequired = errors.New("department name is required")
	ErrJobTitleNameRequired  = errors.New("job title name is required")
	ErrDepartmentRequired    = errors.New("department is required")
)

// validationErrors is the canonical set of "input is invalid" errors. It is
// used by IsValidationError to classify any of them as a 400-level fault
// regardless of the specific field that failed.
var validationErrors = []error{
	ErrFirstNameRequired,
	ErrLastNameRequired,
	ErrEmailRequired,
	ErrEmailInvalid,
	ErrJobTitleRequired,
	ErrCountryRequired,
	ErrSalaryNegative,
	ErrCountryInactive,
	ErrJobTitleInactive,
	ErrDepartmentInactive,
	ErrCountryNameRequired,
	ErrCountryCodeRequired,
	ErrCurrencyRequired,
	ErrDepartmentNameRequired,
	ErrJobTitleNameRequired,
	ErrDepartmentRequired,
}

// IsValidationError reports whether err is one of the validation errors
// declared in this package. Handlers use it to decide whether to return
// HTTP 400 instead of 500.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	for _, target := range validationErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
