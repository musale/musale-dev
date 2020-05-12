---
title: "The Visitor Pattern"
date: 2020-05-12 07:00:22 PM
lastmod: 2020-05-12 07:00:22 PM
authors:
  - Martin Musale
tags:
  - go
  - design patterns
  - visitor pattern
summary: "The [visitor design pattern](/thoughts/visitor-pattern) is about the pattern and also gives an explanation on it's usage in a simple example using Go."
---

# TL;DR

In object-oriented programming and software engineering, the visitor design pattern is a way of separating an algorithm from an object structure on which it operates. A practical result of this separation is the ability to add new operations to existing object structures without modifying the structures.

From [wikipedia](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=41&cad=rja&uact=8&ved=2ahUKEwjZ_K2D4q7pAhVBRBoKHZdPCyEQmhMwKHoECA8QGA&url=https%3A%2F%2Fen.wikipedia.org%2Fwiki%2FVisitor_pattern&usg=AOvVaw1FL-fqGJAkuwqC4C4YZ7YM)

## Introduction

Using an example of cocktails. Say we have a `Cocktail` that we can use to "make" different kinds of cocktails, like `Mojito` and `Daiquiri`. We can have a `Cocktail` interface that defines `Serve` and `Drink` methods for a cocktail:

```go
type Cocktail interface{
    Serve()
    Drink()
}
```

You can make the `Mojito` and `Daiquiri` implement the `Cocktail` interface methods like this:

```go
type Mojito struct{}

func(m Mojito) Serve(){}
func(m Mojito) Drink(){}

type Daiquiri struct{}

func(m Daiquiri) Serve(){}
func(m Daiquiri) Drink(){}
```

### The challenge

Now, we want to be able to define new `Cocktail` operations without having to add the new methods on each existing cocktail. For the above implementation, we'll have to add the method in the `Cocktail` interface and then implement the method for each of the cocktails that we want to implement the method.

The visitor pattern will allow you to operate on a `Cocktail` so that you can provide a `Cocktail` that easily conforms to all the cocktail operations. It makes it easier to add new ways to work on a cocktail without getting to make the change on each and every cocktail.

### The Visitor Pattern

To achieve this, we implement a new interface; `CocktailVisitor` that defines the visitor operations on a cocktail.

```go
type CocktailVisitor interface{
    visitMojito(c Mojito)
    visitDaiquiri(c Daiquiri)
}
```

Next, we need to "route" the correct cocktail to the correct method on the visitor by changing the `Cocktail` interface.

```go
type Cocktail interface{
    accept(v CocktailVisitor)
}
```

With that, we are now able to make the `Cocktail` interface to have a method `accept` that takes in a `CocktailVisitor` or any object "that is" a `CocktailVisitor`.

Finally, we make the cocktails:

```go
type Mojito struct{}
func (m Mojito) accept(v CocktailVisitor){
    v.visitMojito(m)
}

type Daiquiri struct{}
func (d Daiquiri) accept(v CocktailVisitor){
    v.visitDaiquiri(d)
}
```

Now, if you need to add a new cocktail like a `Margharita`, you just have to implement it's visitor and add it to the `CocktailVisitor`.

### Usage

You will need to define the operation you want to accomplish. We want to do a `DrinkVisit` operation:

```go
type DrinkVisit struct{}

func (DrinkVisit) visitMojito(m Mojito) {
	fmt.Println("Drinking mojitos")
}
func (DrinkVisit) visitDaiquiri(d Daiquiri) {
	fmt.Println("Drinking daiquiris")
}
```

Then we do the magic like:

```go
func main() {
	visitor := DrinkVisit{}
	mojito := &Mojito{}
	daiquiri := &Daiquiri{}
	mojito.accept(visitor)
	daiquiri.accept(visitor)
}
```

A working play example can he found [here](https://play.golang.org/p/5f_CjjuCvE-)

### Conclusion

The visitor pattern is a common pattern in language interpeters that allows the interpreter to work on various expressions by defining operations on them. Typically, the `accept` method will return an `interface{}` or a value. That is well beyond the scope of this write up.
