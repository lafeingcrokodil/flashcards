
window.addEventListener("load", () => {
  let isFirstGuess = true;
  let viewCount = 0;
  let correctCount = 0;
  fetchState()
    .then(state => {
      display(state, isFirstGuess, viewCount, correctCount);
      document.querySelector("#answer").addEventListener("keyup", event => {
        if (event.key !== "Enter") return;
        document.querySelector("#submit").click();
        event.preventDefault();
      });
      document.querySelector("#submit").addEventListener("click", () => {
        const answer = document.querySelector("#answer").value;
        submit(answer, isFirstGuess)
          .then(isCorrect => {
            switch (isCorrect) {
              case "true":
                // Update session stats.
                if (isFirstGuess) {
                  correctCount++;
                }
                viewCount++;

                // Reset UI.
                isFirstGuess = true;
                document.querySelector("#answer").value = "";
                document.querySelector("#expected").textContent = "";
                break;
              case "false":
                isFirstGuess = false;
                break;
              default:
                throw new Error(`Invalid submit response: ${isCorrect}`);
            }
          })
          .then(fetchState)
          .then((state) => display(state, isFirstGuess, viewCount, correctCount));
      });
    })
    .catch(err => console.error(err.message));
});

function fetchState() {
  return fetch("state")
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }
      return response.json();
    });
}

function display(state, isFirstGuess, viewCount, correctCount) {
  console.log(state);

  document.querySelector("#unreviewedCount").textContent = state["unreviewed"].length;

  let proficiencyCounts = "";
  for (const count of state["countByProficiency"]) {
    proficiencyCounts += ` · ${count}`;
  }
  document.querySelector("#proficiencyCounts").textContent = proficiencyCounts;

  document.querySelector("#viewCount").textContent = viewCount;
  document.querySelector("#correctCount").textContent = correctCount;
  document.querySelector("#incorrectCount").textContent = viewCount - correctCount;
  document.querySelector("#correctPerc").textContent = percent(correctCount, viewCount);

  const current = state["current"][0];
  document.querySelector("#prompt").textContent = current["prompt"];
  document.querySelector("#context").textContent = current["context"];

  if (!isFirstGuess) {
    document.querySelector("#expected").textContent = current["answers"][0];
  }
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
