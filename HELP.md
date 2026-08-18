# drawdeck

Draw cards at random from markdown checklists.

## Usage

- `drawdeck` — open the deck list
- `drawdeck add FILE.md` — register a deck and open it
- `drawdeck add DIR` — register every markdown file in a folder (`.` for the current one)
- `drawdeck -v` — print version
- `drawdeck -h` — show this help

## List view

- `↑↓` / `jk` — navigate
- `enter` — open deck
- `t` — tournament mode
- `a` — add a deck, or a whole folder of them, by path
- `d` — remove a deck from the list (the file is left alone)
- `r` — re-read every deck file
- `q` — quit

## Deck view

- hold `space` — draw a card
- tap `space` three times — draw a card, on any terminal
- `ctrl+r` — reset the deck
- `esc` — back to the list

## Tournament mode

Runs the whole deck as a single-elimination bracket. `←→` / `hl` pick a winner from each pair until one card is left, recorded as the deck's `winner`. An odd card each round gets a keep-or-discard choice.

Starting a tournament resets the deck, and progress is lost if you leave. A deck with a winner is locked until `ctrl+r`.

## Deck files

The top `#` header names the deck. Every non-indented checkbox is a card; an indented quote beneath it becomes that card's notes.

```
# My Deck

- [ ] A woolly yak
- [ ] A giant elephant
    > This one is my favorite
```

Drawing checks the card off and records it in the file's frontmatter. Nothing else in the file is touched.
