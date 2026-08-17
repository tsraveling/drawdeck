# drawdeck

Draw cards at random from markdown checklists.

## Usage

- `drawdeck` — open the deck list
- `drawdeck add FILE.md` — register a deck and open it
- `drawdeck -v` — print version
- `drawdeck -h` — show this help

## List view

- `↑↓` / `jk` — navigate
- `enter` — open deck
- `a` — add a deck by path
- `d` — remove a deck from the list (the file is left alone)
- `r` — re-read every deck file
- `q` — quit

## Deck view

- hold `space` — draw a card
- tap `space` three times — draw a card, on any terminal
- `ctrl+r` — reset the deck
- `esc` — back to the list

## Deck files

The top `#` header names the deck. Every non-indented checkbox is a card; an indented quote beneath it becomes that card's notes.

```
# My Deck

- [ ] A woolly yak
- [ ] A giant elephant
    > This one is my favorite
```

Drawing checks the card off and records it in the file's frontmatter. Nothing else in the file is touched.
