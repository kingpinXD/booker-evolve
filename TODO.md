# TODO

## Task 1: Fix gofmt and lint issues
**Status:** done
**Result:** gofmt applied to 2 files, 3 errcheck violations fixed. Lint reports 0 issues.

## Task 2: Add FlexDays support to direct search
**Status:** done
**Result:** Added date window filtering via FilterByDateRange when FlexDays > 0. TDD: 2 tests written (flex window, zero-day). direct coverage 94.1% to 94.7%.

## Task 3: Add httpclient unit tests with httptest
**Status:** done
**Result:** 13 tests covering GET/POST, retry logic, 4xx no-retry, BuildURL, connection errors. httpclient coverage 0% to 89.4%.

## Task 4: Add SerpAPI parser unit tests
**Status:** done
**Result:** 11 tests covering all parser functions. parser.go at 100%, package-wide 17.8% to 37.7%.

## Task 5: Add multicity helper unit tests
**Status:** done
**Result:** 22 tests covering deduplicateFlights, buildMultiCityItinerary, fetchFromAllProviders, fetchWithDualSort, NewSearcher, StopoversForRoute. multicity coverage 33.2% to 48.6%.
