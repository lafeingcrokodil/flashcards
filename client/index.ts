window.addEventListener("load", () => {
  let session: Session;
  const sessionId = prompt("Session ID");
  const sessionPromise = !sessionId ? createSession() : getSession(sessionId);
  sessionPromise
    .then((s: Session) => {
      session = s;
      return nextFlashcard(session.id)
    })
    .then((flashcard: Flashcard) => {
      const app = new App(session, flashcard)
      app.display(true);
    })
    .catch((err: Error) => console.error(err.message));
});

class App {
  session: Session;
  flashcard: Flashcard;
  ui: UI;

  isFirstGuess = true;
  viewCount = 0;
  correctCount = 0;

  constructor(session: Session, flashcard: Flashcard) {
    this.session = session;
    this.flashcard = flashcard;
    this.ui = new UI();
    this.ui.answer.addEventListener("keyup", this.handleAnswerKeyup.bind(this));
    this.ui.submit.addEventListener("click", this.handleSubmitClick.bind(this));
    this.ui.allAnswersToggle.addEventListener("click", this.handleAllAnswersToggleClick.bind(this));
  }

  display(isCorrect: boolean) {
    this.ui.unreviewedCount.textContent = this.session.unreviewedCount.toString();

    let proficiencyCounts = "";
    this.session.proficiencyCounts.forEach((count, i) => {
      proficiencyCounts += ` · <span class=${this.getProficiencyClass(i)}>${count}</span>`;
    });
    this.ui.proficiencyCounts.innerHTML = proficiencyCounts;

    this.ui.viewCount.textContent = this.viewCount.toString();
    this.ui.correctCount.textContent = this.correctCount.toString();
    this.ui.incorrectCount.textContent = (this.viewCount - this.correctCount).toString();
    this.ui.correctPerc.textContent = percent(this.correctCount, this.viewCount).toString();

    this.ui.prompt.textContent = this.flashcard.metadata.prompt;
    const context = this.flashcard.metadata.context;
    this.ui.context.textContent = context ? `(${context})` : "";

    if (isCorrect) {
      this.ui.answer.value = "";
    }

    if (this.isFirstGuess) {
      this.ui.expected.textContent = "";
    } else {
      this.ui.expected.textContent = this.flashcard.metadata.answer;
    }

    this.hideAllAnswers();
  }

  getProficiencyClass(repetitions: number): string {
    switch (repetitions) {
      case 0: return "error";
      case 1: return "weak";
      case 2: return "ok";
      case 3: return "strong";
      default: return "correct";
    };
  }

  compareProficiency(a: Flashcard, b: Flashcard): number {
    if (a.stats.viewCount > 0 && b.stats.viewCount == 0) {
      return -1
    }
    if (a.stats.viewCount == 0 && b.stats.viewCount > 0) {
      return 1
    }
    return b.stats.repetitions - a.stats.repetitions
  }

  handleAnswerKeyup(event: KeyboardEvent) {
    if (event.key !== "Enter") return;
    this.ui.submit.click();
    event.preventDefault();
  }

  handleSubmitClick() {
    const answer = this.ui.answer.value;
    submitAnswer(this.session.id, this.flashcard.metadata.id, answer, this.isFirstGuess)
      .then((session: Session | null) => {
        if (session) {
          if (this.isFirstGuess) {
            this.correctCount++;
          }
          this.viewCount++;
          this.isFirstGuess = true;
          this.session = session;
          nextFlashcard(session.id)
            .then((flashcard: Flashcard) => {
              this.flashcard = flashcard;
              this.display(true);
            });
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
    getFlashcards(this.session.id)
      .then((flashcards: Flashcard[]) => {
        flashcards.sort(this.compareProficiency);
        this.ui.allAnswers.innerHTML = " · ";
        for (const f of flashcards) {
          let proficiencyClass = "";
          if (f.stats.viewCount > 0) {
            proficiencyClass = this.getProficiencyClass(f.stats.repetitions);
          }
          this.ui.allAnswers.innerHTML += `<span class=${proficiencyClass}>${f.metadata.answer}</span> · `;
        }
        this.ui.allAnswers.style.display = "block";
        this.ui.allAnswersToggle.value = "▾ Hide answers";
      })
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
    this.prompt = getHTMLElement("#prompt");
    this.context = getHTMLElement("#context");
    this.answer = getHTMLInputElement("#answer");
    this.submit = getHTMLInputElement("#submit");
    this.expected = getHTMLElement("#expected");
    this.allAnswersToggle = getHTMLInputElement("#allAnswersToggle");
    this.allAnswers = getHTMLElement("#allAnswers");
  }
}

interface Session {
  id: string;
  proficiencyCounts: number[];
  unreviewedCount: number;
}

interface Flashcard {
  metadata: FlashcardMetadata;
  stats: FlashcardStats;
}

interface FlashcardMetadata {
  id: number;
  prompt: string;
  context: string;
  answer: string;
}

interface FlashcardStats {
  viewCount: number;
  repetitions: number;
}

async function createSession(): Promise<Session> {
  const response = await fetch(`sessions`, { method: "POST" });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  return response.json();
}

async function getSession(sessionId: string): Promise<Session> {
  const response = await fetch(`sessions/${sessionId}`);
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  return response.json();
}

async function getFlashcards(sessionId: string): Promise<Flashcard[]> {
  const response = await fetch(`sessions/${sessionId}/flashcards`);
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  return response.json();
}

async function nextFlashcard(sessionId: string): Promise<Flashcard> {
  const response = await fetch(`sessions/${sessionId}/flashcards/next`, { method: "POST" });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  return response.json();
}

async function syncFlashcards(sessionId: string): Promise<Session> {
  const response = await fetch(`sessions/${sessionId}/flashcards/sync`, { method: "POST" });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  return response.json();
}

async function submitAnswer(sessionID: string, flashcardID: number, answer: string, isFirstGuess: boolean): Promise<Session | null> {
  const response = await fetch(`sessions/${sessionID}/flashcards/${flashcardID}/submit`, {
    method: "POST",
    body: JSON.stringify({
      "answer": answer,
      "isFirstGuess": isFirstGuess,
    }),
  });

  // The answer was incorrect, so the session wasn't modified.
  if (response.status == 304) {
    return null;
  }

  // The request failed for some reason.
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }

  // The answer was correct and the session was modified.
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
  return Math.floor(100 * numerator / denominator);
}
