
window.addEventListener("load", () => {
  fetchState()
    .then(display)
    .catch((err) => console.error(err.message));
});

function fetchState() {
  return fetch("state")
    .then((response) => {
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }
      return response.json();
    })
}

function display(state) {
    console.log(state);

    document.querySelector("#unreviewedCount").textContent = state["unreviewed"].length;

    let proficiencyCounts = "";
    for (const count of state["countByProficiency"]) {
        proficiencyCounts += ` · ${count}`;
    }
    document.querySelector("#proficiencyCounts").textContent = proficiencyCounts;

    const current = state["current"][0];
    document.querySelector("#prompt").textContent = current["prompt"];
    document.querySelector("#context").textContent = current["context"];
    document.querySelector("#expected").textContent = current["answers"][0];
}
