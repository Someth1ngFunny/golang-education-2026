package main

import "fmt"

func twoSum(nums []int, target int) []int {
	var n int = len(nums)
	var val_ind map[int]int = make(map[int]int)
	for i := range n {
		val_ind[nums[i]] = i
	}
	for i := range n {
		ind, is_contains := val_ind[target-nums[i]]
		if is_contains && i != ind {
			return []int{i, ind}
		}
	}
	return []int{}
}

func main() {
	fmt.Println(twoSum([]int{2, 7, 11, 15}, 9))
}
