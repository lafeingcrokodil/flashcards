window.addEventListener("load", () => {
  getState()
    .then((state: State) => {
      let app = new App(state);
      app.display(true);
    })
    .catch((err: Error) => console.error(err.message));
});

class App {
  state: State;
  ui: UI;

  isFirstGuess = true;
  viewCount = 0;
  correctCount = 0;

  constructor(state: State) {
    this.state = state;
    this.ui = new UI();
    this.ui.pronounce.addEventListener("click", this.handlePronounceClick.bind(this));
    this.ui.answer.addEventListener("keyup", this.handleAnswerKeyup.bind(this));
    this.ui.submit.addEventListener("click", this.handleSubmitClick.bind(this));
    this.ui.allAnswersToggle.addEventListener("click", this.handleAllAnswersToggleClick.bind(this));
  }

  display(isCorrect: boolean) {
    console.log(this.state);

    let unreviewedCount = this.state.unreviewed ? this.state.unreviewed.length : 0;
    for (const f of this.state.current) {
      if (f.viewCount == 0) {
        unreviewedCount++;
      }
    }
    this.ui.unreviewedCount.textContent = unreviewedCount.toString();

    let proficiencyCounts = "";
    this.state.countByProficiency.forEach((count, i) => {
      proficiencyCounts += ` · <span class=${this.getProficiencyClass(i)}>${count}</span>`;
    });
    this.ui.proficiencyCounts.innerHTML = proficiencyCounts;

    this.ui.viewCount.textContent = this.viewCount.toString();
    this.ui.correctCount.textContent = this.correctCount.toString();
    this.ui.incorrectCount.textContent = (this.viewCount - this.correctCount).toString();
    this.ui.correctPerc.textContent = percent(this.correctCount, this.viewCount).toString();

    const current = this.state.current[0];
    if (!current) {
      throw new Error("Current deck is empty");
    }
    this.ui.prompt.textContent = current.prompt;
    const context = current.context;
    this.ui.context.textContent = context ? `(${context})` : "";

    if (isCorrect) {
      this.ui.answer.value = "";
    }

    if (this.isFirstGuess) {
      this.ui.expected.textContent = "";
    } else {
      if (!current.answer) {
        throw new Error("Flashcard is missing an answer");
      }
      this.ui.expected.textContent = current.answer;
    }

    this.hideAllAnswers();
  }

  getProficiencyClass(proficiency: number): string {
    switch (proficiency) {
      case 0: return "error";
      case 1: return "weak";
      case 2: return "ok";
      case 3: return "strong";
      default: return "correct";
    };
  }

  getSortedFlashcards(): Flashcard[] {
    let flashcards = this.state.current.concat(this.state.unreviewed || []);
    for (const deck of this.state.decks) {
      if (deck) {
        flashcards = flashcards.concat(deck);
      }
    }
    flashcards.sort(this.compareProficiency);
    return flashcards
  }

  compareProficiency(a: Flashcard, b: Flashcard): number {
    if (a.viewCount > 0 && b.viewCount == 0) {
      return -1
    }
    if (a.viewCount == 0 && b.viewCount > 0) {
      return 1
    }
    return b.proficiency - a.proficiency
  }

  handlePronounceClick() {
    const current = this.state.current[0];
    if (!current) {
      throw new Error("Current deck is empty");
    }
    new Audio(`audio/${current.answer}.mp3`).play()
      .catch((err: Error) => console.error(err.message));
  }

  handleAnswerKeyup(event: KeyboardEvent) {
    if (event.key !== "Enter") return;
    this.ui.submit.click();
    event.preventDefault();
  }

  handleSubmitClick() {
    const answer = this.ui.answer.value;
    patchState(answer, this.isFirstGuess)
      .then((state: State|null) => {
        if (state) {
          if (this.isFirstGuess) {
            this.correctCount++;
          }
          this.viewCount++;
          this.isFirstGuess = true;
          this.state = state;
          this.display(true);
        } else {
          this.isFirstGuess = false;
          this.display(false);
        }
      });
  }

  handleAllAnswersToggleClick() {
    if (this.ui.allAnswers.style.display === "block") {
      this.hideAllAnswers();
    } else {
      this.displayAllAnswers();
    }
  }

  displayAllAnswers() {
    const sortedFlashcards = this.getSortedFlashcards();
    this.ui.allAnswers.innerHTML = " · ";
    for (const f of sortedFlashcards) {
      let proficiencyClass = "";
      if (f.viewCount > 0) {
        proficiencyClass = this.getProficiencyClass(f.proficiency);
      }
      this.ui.allAnswers.innerHTML += `<span class=${proficiencyClass}>${f.answer}</span> · `;
    }
    this.ui.allAnswers.style.display = "block";
    this.ui.allAnswersToggle.value = "▾ Hide answers";
  }

  hideAllAnswers() {
    this.ui.allAnswers.style.display = "none";
    this.ui.allAnswersToggle.value = "▸ Show all answers by proficiency";
  }
}

class UI {
  unreviewedCount: HTMLElement;
  proficiencyCounts: HTMLElement;
  viewCount: HTMLElement;
  correctCount: HTMLElement;
  incorrectCount: HTMLElement;
  correctPerc: HTMLElement;
  pronounce: HTMLInputElement;
  prompt: HTMLElement;
  context: HTMLElement;
  answer: HTMLInputElement;
  submit: HTMLInputElement;
  expected: HTMLElement;
  allAnswersToggle: HTMLInputElement;
  allAnswers: HTMLElement;

  constructor() {
    this.unreviewedCount = getHTMLElement("#unreviewedCount");
    this.proficiencyCounts = getHTMLElement("#proficiencyCounts");
    this.viewCount = getHTMLElement("#viewCount");
    this.correctCount = getHTMLElement("#correctCount");
    this.incorrectCount = getHTMLElement("#incorrectCount");
    this.correctPerc = getHTMLElement("#correctPerc");
    this.pronounce = getHTMLInputElement("#pronounce");
    this.prompt = getHTMLElement("#prompt");
    this.context = getHTMLElement("#context");
    this.answer = getHTMLInputElement("#answer");
    this.submit = getHTMLInputElement("#submit");
    this.expected = getHTMLElement("#expected");
    this.allAnswersToggle = getHTMLInputElement("#allAnswersToggle");
    this.allAnswers = getHTMLElement("#allAnswers");
  }
}

interface State {
  current: Flashcard[];
  unreviewed: Flashcard[];
  decks: Flashcard[][];
  countByProficiency: number[];
}

interface Flashcard {
  prompt: string;
  context: string;
  answer: string;
  viewCount: number;
  proficiency: number;
}

async function getState(): Promise<State> {
  const response = await fetch("state");
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  return response.json();
}

async function patchState(answer: string, isFirstGuess: boolean): Promise<State|null> {
  const response = await fetch("state", {
    method: "PATCH",
    body: JSON.stringify({
      "answer": answer,
      "isFirstGuess": isFirstGuess,
    }),
  });

  // The answer was incorrect, so the state wasn't modified.
  if (response.status == 304) {
    return null;
  }

  // The request failed for some reason.
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }

  // The answer was correct and the state was modified.
  return response.json();
}

function getHTMLElement(selector: string): HTMLElement {
  const elem = document.querySelector(selector);
  if (!elem) {
    throw new Error(`Element not found: ${selector}`);
  }
  return elem as HTMLElement;
}

function getHTMLInputElement(selector: string): HTMLInputElement {
  const elem = getHTMLElement(selector);
  return elem as HTMLInputElement;
}

function percent(numerator: number, denominator: number): number {
  if (denominator === 0) {
    return 0;
  }
  return Math.floor(100*numerator/denominator);
}
