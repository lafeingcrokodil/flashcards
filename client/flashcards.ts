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

    this.ui.unreviewedCount.textContent = this.state["unreviewed"].length.toString();

    let proficiencyCounts = "";
    this.state["countByProficiency"].forEach((count, i) => {
      proficiencyCounts += ` · <span class=${this.ui.getProficiencyClass(i)}>${count}</span>`;
    });
    this.ui.proficiencyCounts.innerHTML = proficiencyCounts;

    this.ui.viewCount.textContent = this.viewCount.toString();
    this.ui.correctCount.textContent = this.correctCount.toString();
    this.ui.incorrectCount.textContent = (this.viewCount - this.correctCount).toString();
    this.ui.correctPerc.textContent = percent(this.correctCount, this.viewCount).toString();

    const current = this.state["current"][0];
    this.ui.prompt.textContent = current["prompt"];
    const context = current["context"];
    this.ui.context.textContent = context ? `(${context})` : "";

    if (isCorrect) {
      this.ui.answer.value = "";
    }

    if (this.isFirstGuess) {
      this.ui.expected.textContent = "";
    } else {
      this.ui.expected.textContent = current["answers"][0];
    }
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
      })
  }

  updateState() {
    return getState()
      .then((state: State) => this.state = state);
  }
}

class UI {
  unreviewedCount: Element;
  proficiencyCounts: Element;
  viewCount: Element;
  correctCount: Element;
  incorrectCount: Element;
  correctPerc: Element;
  prompt: Element;
  context: Element;
  answer: HTMLInputElement;
  submit: HTMLInputElement;
  expected: Element;

  constructor() {
    this.unreviewedCount = getElement("#unreviewedCount");
    this.proficiencyCounts = getElement("#proficiencyCounts");
    this.viewCount = getElement("#viewCount");
    this.correctCount = getElement("#correctCount");
    this.incorrectCount = getElement("#incorrectCount");
    this.correctPerc = getElement("#correctPerc");
    this.prompt = getElement("#prompt");
    this.context = getElement("#context");
    this.answer = getHTMLInputElement("#answer");
    this.submit = getHTMLInputElement("#submit");
    this.expected = getElement("#expected");
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

function getState(): Promise<State> {
  return fetch("state")
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }
      return response.json();
    });
}

function patchState(answer: string, isFirstGuess: boolean): Promise<State|null> {
  return fetch("state", {
    method: "PATCH",
    body: JSON.stringify({
      "answer": answer,
      "isFirstGuess": isFirstGuess,
    }),
  })
    .then(response => {
      // The answer was incorrect, so the state wasn't modified.
      if (response.status == 304) {
        return null
      }

      // The request failed for some reason.
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }

      // The answer was correct and the state was modified.
      return response.json();
    });
}

function getElement(selector: string): Element {
  const elem = document.querySelector(selector);
  if (!elem) {
    throw new Error(`Element not found: ${selector}`);
  }
  return elem;
}

function getHTMLInputElement(selector: string): HTMLInputElement {
  const elem = getElement(selector);
  return elem as HTMLInputElement;
}

function percent(numerator: number, denominator: number): number {
  if (denominator === 0) {
    return 0;
  }
  return Math.trunc(100*numerator/denominator);
}
