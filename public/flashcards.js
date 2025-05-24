window.addEventListener("load", () => {
  fetchState()
    .then(state => {
      let app = new App(state);
      app.display();
    })
    .catch(err => console.error(err.message));
});

class App {
  constructor(state) {
    this.state = state;
    this.isFirstGuess = true;
    this.viewCount = 0;
    this.correctCount = 0;
    this.ui = new UI();
    this.ui.answer.addEventListener("keyup", this.handleAnswerKeyup.bind(this));
    this.ui.submit.addEventListener("click", this.handleSubmitClick.bind(this));
  }

  display() {
    console.log(this.state);

    this.ui.unreviewedCount.textContent = this.state["unreviewed"].length;

    let proficiencyCounts = "";
    this.state["countByProficiency"].forEach((count, i) => {
      proficiencyCounts += ` · <span class=${this.ui.getProficiencyClass(i)}>${count}</span>`;
    });
    this.ui.proficiencyCounts.innerHTML = proficiencyCounts;

    this.ui.viewCount.textContent = this.viewCount;
    this.ui.correctCount.textContent = this.correctCount;
    this.ui.incorrectCount.textContent = this.viewCount - this.correctCount;
    this.ui.correctPerc.textContent = percent(this.correctCount, this.viewCount);

    const current = this.state["current"][0];
    this.ui.prompt.textContent = current["prompt"];
    const context = current["context"];
    this.ui.context.textContent = context ? `(${context})` : "";

    if (!this.isFirstGuess) {
      this.ui.expected.textContent = current["answers"][0];
    }
  }

  handleAnswerKeyup(event) {
    if (event.key !== "Enter") return;
    this.ui.submit.click();
    event.preventDefault();
  }

  handleSubmitClick() {
    const answer = this.ui.answer.value;
    submit(answer, this.isFirstGuess)
      .then(isCorrect => {
        switch (isCorrect) {
          case "true":
            if (this.isFirstGuess) {
              this.correctCount++;
            }
            this.viewCount++;
            this.isFirstGuess = true;
            this.ui.reset();
            break;
          case "false":
            this.isFirstGuess = false;
            break;
          default:
            throw new Error(`Invalid submit response: ${isCorrect}`);
        }
      })
      .then(this.updateState.bind(this))
      .then(this.display.bind(this));
  }

  updateState() {
    return fetchState()
      .then(state => this.state = state);
  }
}

class UI {
  constructor() {
    this.unreviewedCount = document.querySelector("#unreviewedCount");
    this.proficiencyCounts = document.querySelector("#proficiencyCounts");
    this.viewCount = document.querySelector("#viewCount");
    this.correctCount = document.querySelector("#correctCount");
    this.incorrectCount = document.querySelector("#incorrectCount");
    this.correctPerc = document.querySelector("#correctPerc");
    this.prompt = document.querySelector("#prompt");
    this.context = document.querySelector("#context");
    this.answer = document.querySelector("#answer");
    this.submit = document.querySelector("#submit");
    this.expected = document.querySelector("#expected");
  }

  getProficiencyClass(proficiency) {
    switch (proficiency) {
      case 0: return "error";
      case 1: return "weak";
      case 2: return "ok";
      case 3: return "strong";
      default: return "correct";
    };
  }

  reset() {
    this.answer.value = "";
    this.expected.textContent = "";
  }
}

function fetchState() {
  return fetch("state")
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }
      return response.json();
    });
}

function submit(answer, isFirstGuess) {
  return fetch("submit?" + new URLSearchParams({
    "answer": answer,
    "isFirstGuess": isFirstGuess,
  }))
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }
      return response.text();
    });
}

function percent(numerator, denominator) {
  if (denominator === 0) {
    return 0;
  }
  return Math.trunc(100*numerator/denominator);
}
