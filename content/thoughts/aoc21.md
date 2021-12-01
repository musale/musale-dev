---
title: "Advent of Code 2021"
date: 2021-12-01T08:36:14+03:00
lastMod: 2021-12-01T08:36:40:14+03:00
authors:
  - Martin Musale
tags:
  - Advent of Code 2021
  - Golang
  - Python3
  - Algorithms
  - Data Structures
keywords:
  - Advent of Code 2021
  - Golang
  - Python3
  - Algorithms
  - Data Structures
description: "[The Advent of Code 2019](/thoughts/aoc21/) has my thoughts about the daily [exercises](https://adventofcode.com/) which I'm currently attempting to solve with both _Go_ and _Python 3_."
---

# Motivation

I always like solving problems and this is a very nice place to practise my skills. This year, I would like to learn more concepts and I'm already liking this _AOC_. For starters, we have a lot of participation and people are streaming when they're solving the puzzles. I like the way [Jonathan Paulson's](https://www.youtube.com/channel/UCuWLIm0l4sDpEe28t41WITA/videos) videos are done. They are precise and really nice to see. He's fast too and that makes me want to look and think through a puzzle with his level of understanding and speed. I also like [Liz Fong-Jones'](https://www.twitch.tv/lizthegrey) _Go_ solutions. My solutions are over [here](https://github.com/musale/advent-of-code-2021).


## Day 1

Pretty simple exercise. It involved getting the number of times values in an array increase. So given an array like `[1,2,3,2,4,1]`, the increases are 3. In the first part, the task was to get the total number of increases going through the array. In the second part, the task was switched up to get the sum of increases in a sliding window of three values. So for `[1,2,3,2,4,1]` the first window has `[1,2,3]` and the next window has `[2,3,2]`. If you sum that, then there is an increase and that's the first one. You go on until the end of the array where the values in the window are less than three.

You can find my solution to this exercise in *golang* [here.](https://github.com/musale/advent-of-code-2021/blob/master/day_01/go/main.go)