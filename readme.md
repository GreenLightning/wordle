## Cheating at Wordle

You might know that Wordle is a client-side game and picks the target word based
on the current date, so that everyone plays the same game on any given day.
Therefore it would be trivial to reverse engineer the code and compute the
solution, allowing you to win in one guess every time. So, if we are going to
cheat, we have to decide what is allowed. This program provides several options.

### Word Lookup

By default the program acts as a dictionary, listing all the words that match a
given set of hints. Hints are a compressed form of the clues you get from the
game. A hint string always starts with 5 letters or underscores, representing
the fixed letters in the word. You can require or exclude letters using `+` and
`.`. Here are some examples:

All words starting and ending with an `A`:

```
>wordle A___A
AGORA
ALPHA
AORTA
APNEA
ARENA
AROMA
```

All words containing 3 or more `T`s:

```
>wordle _____+TTT
TATTY
```

All words not containing any vowel:

```
>wordle _____-AEIOU
CRYPT
DRYLY
GLYPH
GYPSY
LYMPH
LYNCH
MYRRH
NYMPH
PYGMY
SHYLY
SLYLY
TRYST
WRYLY
```

You can decide for yourself how many hints you want to pass to the program.
Or maybe you just want to see all the words with a `T` in the second position
to choose one for your next guess.

Here are some more advanced usages regarding letter count and positioning:

If you put a letter in both the fixed/required and excluded sections,
that means the count has to be exact:

```
> wordle _____+AA   # at least two As
> wordle A____+A    # at least two As, one of which must be the first letter
> wordle _____+AA-A # exactly two As, in any position
```

If you get a yellow clue, that actually means the target word does not contain
this letter at this position (or else it would have been marked green).
You can express this by putting one or more numbers after a required letter:

```
> wordle _____+A1   # must contain an A but not as the first letter
> wordle _____+AA1  # must contain at least two As but not as the first letter
> wordle _____+A345 # must contain an A but not as one of the last three letters
```

### Best Starting Words

I was also interested in finding the best starting word(s). My metric for rating
a guess is how many other words it eliminates on average. You can search for the
best word using `wordle -best` and for the best two words using `wordle -best2`.
For a single word all possible input words are considered. For two words only
the smaller list of possible target words is considered. This still takes a
while (~22 minutes) and is a great way to use your CPU to heat your room. For
your convenience, the results are [here](results.md) (trying to avoid spoilers
in this readme).

The percentage after each guess indicates how many words remain on average. For
example `TOAST 7.178%` means that after guessing `TOAST` the clues will allow
you to rule out all but 166 of 2315 possible words. This is of course only an
average. You can see the distribution using `-dist`:

```
TOAST 7.178%
  0 ***************************************************************************************
 50 ***********************
100 *******************************************************************
150
200
250 *****************************
300
350
400 *************************************************
450
```

So in most cases you will have less than 50 words left to decide between, but
there are some outliers where you will still have more than 400 words left.

### Best Next Guess

You can also combine both modes and compute the best next guess for a set of hints.

For example, if you know the word starts with `P` and ends with `A`, there are
five possible words that fit:

```
>wordle P___A
PARKA
PASTA
PIZZA
PLAZA
POLKA
```

And the best guess is `PARKA`:

```
>wordle -for P___A
5
PARKA 20.000%
ABACK 20.000%
ALARM 20.000%
```

The `5` indicates how many words match the hints, then the best guesses are
listed in order.

`PARKA` is the best guess and might actually be correct.
If not, and the first `A` is green, then it must be `PASTA`.
Otherwise, if the first `A` is yellow, it must be `PLAZA`.
Otherwise, if the `K` is green, it must be `POLKA`.
And if none of the above apply, it must be `PIZZA`.
