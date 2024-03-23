let f: HTMLFormElement = document.querySelector('form');
let l = document.getElementById('loading');
f.addEventListener('submit', (e) => {
	e.preventDefault();
	f.hidden = true;
	l.hidden = false;
	fetch("/api/add?etkn=95d9d334b7dc7fd211b3", {
		method: "POST",
		headers: {
			"Content-Type": "application/json"
		},
		body: JSON.stringify({ name: (document.getElementById('fname') as HTMLInputElement).value,
		lname: (document.getElementById('lname') as HTMLInputElement).value, 
		email: (document.getElementById('eml') as HTMLInputElement).value, 
		msg: (document.getElementById('msg') as HTMLTextAreaElement).value
		}),
	}).then((s) => {
		l.hidden = true;
		f.hidden = false;
		if (s.status === 200) {
			f.reset();
			return alert("Message sent successfully\nWe will get back to you soon!");
		}
		console.error(s);
		alert("An error occured while sending the message\nPlease try again. We apologize for the inconvenience.");
	});
});
console.info("Website made by 1Alphabyte (https://git.utsav2.dev/utsav)");