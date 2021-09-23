package calldata

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
)

const (
	minLen         = 2
	maxLen         = 64
	defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Func struct {
	Name string
	Desc string
	Func interface{}
}

var (
	funcs = []Func{
		{
			"randomNumber",
			"Generate a number in range [min=0, max)",
			randomdata.Number,
		},
		{
			"randomDecimal",
			"Generate a number in range [min=0, max) with decimal point x",
			randomdata.Decimal,
		},
		{
			"randomBoolean",
			"Generate a bool",
			randomdata.Boolean,
		},
		{
			"randomString",
			"Generate a string from a-zA-Z0-9. Accepts a length parameter",
			randomString,
		},
		{
			"randomStringWithCharset",
			"Generate a string from charset",
			randomStringWithCharset,
		},
		{
			"randomStringSample",
			"Generate a string sampled from a list of strings",
			randomdata.StringSample,
		},
		{
			"randomSillyName",
			"Generate a silly name",
			randomdata.SillyName,
		},
		{
			"randomMaleName",
			"Generate a male name",
			func() string {
				return randomdata.FirstName(randomdata.Male)
			},
		},
		{
			"randomFemaleName",
			"Generate a female name",
			func() string {
				return randomdata.FirstName(randomdata.Female)
			},
		},
		{
			"randomName",
			"Generate a human name",
			func() string {
				return randomdata.FirstName(randomdata.RandomGender)
			},
		},
		{
			"randomMaleFullName",
			"Generate a male full name",
			func() string {
				return randomdata.FullName(randomdata.Male)
			},
		},
		{
			"randomFemaleFullName",
			"Generate a female full name",
			func() string {
				return randomdata.FullName(randomdata.Female)
			},
		},
		{
			"randomFullName",
			"Generate a human full name",
			func() string {
				return randomdata.FullName(randomdata.RandomGender)
			},
		},
		{
			"randomEmail",
			"Generate a email",
			randomdata.Email,
		},
		{
			"randomIpV4Address",
			"Generate a valid random IPv4 address",
			randomdata.IpV4Address,
		},
		{
			"randomIpV6Address",
			"Generate a valid random IPv6 address",
			randomdata.IpV6Address,
		},
		{
			"randomDateInRange",
			"Generate a full date in range",
			randomdata.FullDateInRange,
		},
		{
			"randomPhoneNumber",
			"Generate a phone number",
			randomdata.PhoneNumber,
		},
		{
			"string",
			"Return the string representation of argument",
			func(arg interface{}) string {
				return fmt.Sprintf("%v", arg)
			},
		},
		{
			"quote",
			"Return a double-quoted string literal representing string",
			strconv.Quote,
		},
		{
			"uuid",
			"Generate a new UUID",
			uuid.NewString,
		},
	}

	random *rand.Rand
)

// SetSeeder sets seed for pseudo-random generator.
func SetSeeder(seed int64) {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	random = rand.New(rand.NewSource(seed)) //nolint:gosec
	randomdata.CustomRand(random)
}

func randomStringWithCharset(charset string, length int) string {
	if length <= 0 {
		length = randomdata.Number(maxLen-minLen+1) + minLen
	}
	if charset == "" {
		charset = defaultCharset
	}

	runes := []rune(charset)
	count := len(runes)

	data := make([]rune, 0, length)
	for i := 0; i < length; i++ {
		data = append(data, runes[randomdata.Number(count)])
	}

	return string(data)
}

func randomString(length int) string {
	return randomStringWithCharset(defaultCharset, length)
}
