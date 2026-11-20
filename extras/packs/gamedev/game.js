const canvas = document.getElementById("game");
const ctx = canvas.getContext("2d");
const keys = new Set();
const player = new Player(40, 40);
let last = performance.now();

addEventListener("keydown", (event) => {
  if (event.code.startsWith("Arrow") || event.code.startsWith("Key")) keys.add(event.code);
});
addEventListener("keyup", (event) => keys.delete(event.code));

function frame(now) {
  const dt = Math.min(0.05, (now - last) / 1000);
  last = now;
  player.update(dt, keys, { w: canvas.width, h: canvas.height });

  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.strokeStyle = "#252a35";
  for (let x = 0; x < canvas.width; x += 40) {
    ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, canvas.height); ctx.stroke();
  }
  for (let y = 0; y < canvas.height; y += 40) {
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(canvas.width, y); ctx.stroke();
  }
  player.draw(ctx);
  requestAnimationFrame(frame);
}

requestAnimationFrame(frame);
