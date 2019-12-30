---
title: "Advent of Code 2019"
date: 2019-12-06T20:09:16+03:00
lastmod: 2019-12-06T20:09:16+03:00
authors:
  - Martin Musale
tags:
  - Advent of Code
  - Golang
  - Python3
  - Algorithms
  - Data Structures
summary: "[The Advent of Code 2019](/thoughts/aoc19/) has my thoughts about the daily [exercises](https://adventofcode.com/) which I'm currently attempting to solve with both _Go_ and _Python 3_."
---

# Motivation

I always like solving problems and this is a very nice place to practise my skills. This year, I would like to learn more concepts and I'm already liking this _AOC_. For starters, we have a lot of participation and people are streaming when they're solving the puzzles. I like the way [Jonathan Paulson's](https://www.youtube.com/channel/UCuWLIm0l4sDpEe28t41WITA/videos) videos are done. They are precise and really nice to see. He's fast too and that makes me want to look and think through a puzzle with his level of understanding and speed. I also like [Liz Fong-Jones'](https://www.twitch.tv/lizthegrey) _Go_ solutions. My solutions are over [here](https://github.com/musale/advent-of-code-2019).

## Day 9

The `IntCode` makes a come back but we are prepared :fork_and_knife:! I liked this puzzle because in as much as you needed to rethink how you fetch the modes, that was all there was. In the second part it was a matter of changing the inputs and boom! Very easy and almost, just the way I like it.

For me, I have the most of this `IntCode` solutions in Python. What I would like to do is to port it to `Go`.

## Day 8

Decoding the image data was pretty easy. It involved array manipulation so from a higher level you can see the kind of problem that you have and the way the ideal solution needs to look like. For my case, I knew that I need a way to keep layers, which I did using a `dict`. To find the layer with the most common `zeros`, I used Python's `collections.Counter` to find the count of value 0 and set the lowest value for each layer. Pretty easy.

For the second part, I looped through the size of an image (`width * height`) and then the value which is not 2 in a second loop through the layers. This enabled me to look at the previous layer data as I went through looking for a pixel value that's not 2.

Printing the value to get what text it looks like was an awesome task. I got this:

    ####.###..####.#..#.###..
    #....#..#....#.#..#.#..#.
    ###..###....#..#..#.#..#.
    #....#..#..#...#..#.###..
    #....#..#.#....#..#.#.#..
    ####.###..####..##..#..#.

Looking towards day #9!

## Day 7

**Amplifiers**! This puzzle builds up from the [day 5](https://adventofcode.com/2019/day/5) `IntCode` puzzle. There's multi processing in part 2 which really had me in a mind blown state. For some reason ~~which I don't know~~ I just find it difficult understanding the int code puzzles. I managed to do the first part but translating that to part 2 became a bit of a challenge. One thing I learnt how to tackle it was to make the code I'm running more clear. I came across a solution by [alcatrazEscapee](https://github.com/alcatrazEscapee/AdventofCode/blob/master/2019/day7/day7.py) and reading through his code made me understand the puzzle even better. I hope day 8 is easy on me :laugh:.

## Day 6

Yay! A simple puzzle. We had ourselves a tree and traversals. This was a bit easy to do. Also, the illustrations make it really easy to think through your solutions.

## Day 5

**Good Lord!** This is by far the most difficult of the puzzles. For starters, it uses day 2's code. Then the theory is just a lot. My brain got lost for a couple of hours here. I finally followed Jonathan Paulson's steps and managed to finesse it. Sigh

## Day 1 - 4

These exercises were pretty simple and straightforward. [Day 1](https://adventofcode.com/2019/day/1) exercises involved summing values, pretty easy. [Day 2](https://adventofcode.com/2019/day/2) had some splitting values given a certain step and then doing different arithmetic and end up with a final value. That was pretty easy to do too.

[Day 3](https://adventofcode.com/2019/day/3) came with Manhattan distance. Owh, this one was nice... I'm still unable to do the visualizations and it's one thing that I found out people are doing with their solutions. It's fantastic seeing people doing these.

[Day 4](https://adventofcode.com/2019/day/4) was _brute force_ day. Genarating values in a given range and using different criteria to determine the values you need and then make a final value.
