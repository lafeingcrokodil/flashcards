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
    this.ui.answer.addEventListener("keyup", this.handleAnswerKeyup.bind(this));
    this.ui.submit.addEventListener("click", this.handleSubmitClick.bind(this));
  }

  display(isCorrect: boolean) {
    console.log(this.state);

    this.ui.unreviewedCount.textContent = this.state.unreviewed.length.toString();

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
      const expected = current.answers[0];
      if (!expected) {
        throw new Error("Flashcard is missing answers");
      }
      this.ui.expected.textContent = expected;
    }
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
}

class UI {
  unreviewedCount: HTMLElement;
  proficiencyCounts: HTMLElement;
  viewCount: HTMLElement;
  correctCount: HTMLElement;
  incorrectCount: HTMLElement;
  correctPerc: HTMLElement;
  prompt: HTMLElement;
  context: HTMLElement;
  answer: HTMLInputElement;
  submit: HTMLInputElement;
  expected: HTMLElement;

  constructor() {
    this.unreviewedCount = getHTMLElement("#unreviewedCount");
    this.proficiencyCounts = getHTMLElement("#proficiencyCounts");
    this.viewCount = getHTMLElement("#viewCount");
    this.correctCount = getHTMLElement("#correctCount");
    this.incorrectCount = getHTMLElement("#incorrectCount");
    this.correctPerc = getHTMLElement("#correctPerc");
    this.prompt = getHTMLElement("#prompt");
    this.context = getHTMLElement("#context");
    this.answer = getHTMLInputElement("#answer");
    this.submit = getHTMLInputElement("#submit");
    this.expected = getHTMLElement("#expected");
  }
}

interface State {
  unreviewed: Flashcard[];
  countByProficiency: number[];
  current: Flashcard[];
}

interface Flashcard {
  prompt: string;
  context: string;
  answers: string[];
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
