const out = document.getElementById("out");
let clicks = 0;

document.getElementById("go").addEventListener("click", () => {
  clicks += 1;
  out.textContent = `Clicked ${clicks} time${clicks === 1 ? "" : "s"}.`;
});
