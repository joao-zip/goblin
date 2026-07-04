package demo

func IsAdult(age int) bool {
	return age >= 18
}

func Discount(price float64, percentage float64) float64 {
	return price - (price * percentage / 100)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func IsEven(n int) bool {
	return n%2 == 0
}

func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
