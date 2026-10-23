document.addEventListener('click', (event) => {
  const link = event.target.closest('a[data-nsfw]');
  if (link && !window.confirm('18+ WARNING — This board may contain NSFW content. Are you 18+ and want to continue?')) event.preventDefault();
});

// Count scalars to match the server rather than UTF-16 string length.
const body = document.querySelector('textarea[name=body]');
const bodyCount = document.querySelector('.body-count');
if (body && bodyCount) {
  const update = () => {
    const n = [...body.value].length;
    bodyCount.value = `${n} / 5000 characters${n > 5000 ? ' (too long)' : ''}`;
  };
  bodyCount.hidden = false;
  body.addEventListener('input', update);
  update();
}

const image = document.querySelector('input[name=image]');
if (image) {
  document.addEventListener('paste', (event) => {
    const file = [...event.clipboardData.files].find((f) => image.accept.split(',').includes(f.type));
    if (!file) return;
    const files = new DataTransfer();
    files.items.add(file);
    image.files = files.files;
  });
}

const relative = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
const units = [['year', 31536000], ['month', 2592000], ['day', 86400], ['hour', 3600], ['minute', 60], ['second', 1]];
for (const time of document.querySelectorAll('time[datetime]')) {
  const seconds = (Date.parse(time.dateTime) - Date.now()) / 1000;
  const [unit, size] = units.find(([, size]) => Math.abs(seconds) >= size) ?? units.at(-1);
  time.title = time.textContent;
  time.textContent = relative.format(Math.round(seconds / size), unit);
}

const corner = document.querySelector('.corner-img');
if (corner) {
  const images = ['/icons/corner.png', '/icons/corner2.png', '/icons/corner3.png'];
  corner.src = images[Math.floor(Math.random() * images.length)];
}
