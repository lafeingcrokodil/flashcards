# Vocab trainer prototype

A single-page Korean vocabulary and reading trainer, built as a Claude
artifact, to try out spaced repetition and reading practice on a phone.
Everything is in [`index.html`](index.html): markup, styles and script, with no
build step.

It has four tabs:

- **Cards:** Korean → English flashcards. A session is up to 15 cards: due
  reviews first, then a chosen number of new words (5, 8 or 12).
- **Type:** English → Korean recall. The learner types the dictionary form,
  with an optional first-syllable hint and an "I was right" override for
  answers the exact match rejects.
- **Write:** the learner writes a sentence with a word they've seen, and
  Claude marks it correct, a small fix or needs work, with a corrected
  sentence and a short explanation. The five latest sentences are listed.
- **Read:** short passages (dialogues, a vlog, fan comments, a news text),
  translated line by line with Claude's feedback and a reference translation,
  or answered as multiple-choice questions. Tapping a word shows its
  dictionary form, meaning and any attached ending, from the passage's glossary
  or, failing that, from Claude. A new word can be added to a "My saved words"
  pack, always as its dictionary form with its dictionary meaning. A word
  already studied can be sent back to box 0 and made due now ("Review again"),
  without counting as a miss. Clicking outside the popup closes it.

Scheduling is a Leitner system shared by Cards and Type: a right answer moves
a word up one box (due in 1, 2, 4, 8, 16 or 32 days), a wrong one sends it back
to box 0 and repeats it once at the end of the session. A word counts as
learned from box 3.

Cards are in dictionary form (먹다, not 먹어요). Spoken forms such as 괜찮아
go in a card's alternative answers and note.

Words can be filtered by pack. A pack is a list of words, each with Korean,
English, an example sentence and its translation, an optional note and
optional alternative answers for Type.

## Data

The words and passages aren't in the page: they're documents in the
artifact's database, written by Claude.

| Collection     | One document per | Holds                                                         |
| -------------- | ---------------- | ------------------------------------------------------------- |
| `packs`        | pack             | title, order and the words                                    |
| `passages`     | passage          | title, level, focus, lines, questions and a glossary by token¹ |
| `progress`     | word             | box, due time, right and wrong counts, last review            |
| `translations` | passage line     | the latest attempt, verdict and feedback, and earlier attempts |
| `quiz`         | passage          | answers and score                                             |
| `sentences`    | sentence         | the word, the sentence and Claude's correction                |

¹ Each glossary entry has `d` (dictionary form), `de` (dictionary meaning,
when it differs from the meaning in context), `en` (meaning in this sentence)
and `n` (a note on any attached ending or particle).

## Backups

The data lives only in the artifact's database, so the page's footer offers
two ways to export it, both producing the same JSON file:

- **Back up to Drive** uploads `korean-vocab-backup-<YYYY-MM-DD_HH-MM>.json`
  (UTC) to a "Korean Vocab Backups" folder in the viewer's Google Drive. The
  folder is found by its exact name and created if it's missing. Each backup
  is a new file; older ones are never overwritten.
- **Download backup (JSON)** offers the same file as a download.

The file is meant to be imported into the app later:

```json
{
  "format": "korean-vocab-trainer-export",
  "version": 1,
  "exportedAt": "2026-11-21T12:30:00.000Z",
  "source": "https://claude.ai/artifact/DiymVuAMXqkvH7GwEDzK16",
  "collections": {
    "packs": { "<doc id>": { "...": "document fields as stored" } },
    "progress": {},
    "passages": {},
    "translations": {},
    "quiz": {},
    "sentences": {}
  }
}
```

Each collection from the table above maps document IDs to their fields,
unchanged. Progress and sentences are personal, so backups belong in private
storage, never in this repository.

## Running it

The page expects the artifact runtime's `window.claude.use()`: `db` for the
data above, `sample` for Claude's feedback and word lookups, `tts` for
hearing words in a Claude voice, which isn't offered in Safari or the mobile
apps, `downloads` for saving a backup file, and `mcp` for backing up to
Google Drive through the viewer's Google Drive connector (only its
`search_files` and `create_file` tools). Without any of these, the features
that need it are hidden. Opened as a plain
file, without the runtime, there's no data to load, so the page only says so.
