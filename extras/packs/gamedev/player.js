class Player {
  constructor(x, y) {
    this.x = x;
    this.y = y;
    this.size = 22;
    this.speed = 220;
  }

  update(dt, keys, bounds) {
    let dx = 0;
    let dy = 0;
    if (keys.has("ArrowLeft") || keys.has("KeyA")) dx -= 1;
    if (keys.has("ArrowRight") || keys.has("KeyD")) dx += 1;
    if (keys.has("ArrowUp") || keys.has("KeyW")) dy -= 1;
    if (keys.has("ArrowDown") || keys.has("KeyS")) dy += 1;
    if (dx && dy) { dx *= Math.SQRT1_2; dy *= Math.SQRT1_2; }
    this.x = Math.max(0, Math.min(bounds.w - this.size, this.x + dx * this.speed * dt));
    this.y = Math.max(0, Math.min(bounds.h - this.size, this.y + dy * this.speed * dt));
  }

  draw(ctx) {
    ctx.fillStyle = "#5eead4";
    ctx.fillRect(this.x, this.y, this.size, this.size);
  }
}
