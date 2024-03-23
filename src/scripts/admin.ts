let le = document.getElementById('loading');
async function fetchData() {
	le.hidden = false;
	const res = await fetch('/api/get', {
		mode: 'same-origin',
	});
	const h1: HTMLHeadingElement = document.querySelector('div h1');
	if (res.status == 204) {
		le.hidden = true;
		h1.innerText = 'No data found';
		return;
	} else if (res.status === 403 || res.status === 401) {
		document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/api;';
		localStorage.removeItem('tokenExists');
		localStorage.removeItem('tokenExpiry');
		return location.reload();
	} else if (!res.ok) {
		le.hidden = true;
		h1.innerText = 'An unexpected error occurred.\n Please try again later or contact support.';
		console.error(res);
		return;
	}
	const data = await res.json();
	data.forEach((item) => {
		const row = document.createElement('tr');
		let key = item.key;
		item = item.value;
		row.innerHTML = `
		<td>${item.name} ${item.lname ?? ""}</td>
		<td><a href="mailto:${item.email}">${item.email}</a></td>
		<td>${item.msg}</td>
		<td data-key=${key} class=del>Delete</td>
		`;
		document.querySelector('tbody').appendChild(row);
	});
	le.hidden = true;
	document.querySelector('table').hidden = false;
	document.querySelectorAll('.del').forEach((el) => {
		el.addEventListener('click', async (e) => {
			const key = (e.target as HTMLTableCellElement).dataset.key;
			const r = await fetch(`/api/delete`, {
				method: 'DELETE',
				body: key,
			});
			if (r.ok) {
				(e.target as HTMLTableCellElement).parentElement.remove();
				if (document.querySelectorAll('tbody tr').length === 0) {
					return location.reload();
				}
			}
		});
	});
}

document.getElementById("login-form").addEventListener("submit", (e) => {
	e.preventDefault();
	let f = document.getElementById("login-form") as HTMLFormElement;
	let l = document.getElementById("loading");
	f.hidden = true;
	l.hidden = false;
	fetch("/api/signin", {
		method: "POST",
		mode: "same-origin",
		headers: {
			"Content-Type": "application/json"
		},
		body: JSON.stringify({ 
				username: (document.getElementById('username') as HTMLInputElement).value,
				password: (document.getElementById('password') as HTMLInputElement).value
			}
		),
	}).then(async (s) => {
		l.hidden = true;
		f.hidden = false;
		if (s.status === 200) {
			f.reset();
			document.getElementById("main").hidden = false;
			document.querySelector('body').removeChild(f)
			let b = await s.json();
			console.log(b);
			document.cookie = `token=${b.sessionToken}; path=/api; samesite=strict; expires=${new Date(b.expiresAt * 1000).toUTCString()}; secure`;
			localStorage.setItem('tokenExists', 'true');
			localStorage.setItem('tokenExpiry', new Date(b.expiresAt * 1000).toString());
			return fetchData();
		} else if (s.status === 403) {
			return alert("Invalid username or password");
		}
		console.error(s);
		alert("An error occured while logging in\nPlease try again. We apologize for the inconvenience.");
	});
});

if (localStorage.getItem('tokenExists') === 'true' && new Date(localStorage.getItem('tokenExpiry')) > new Date()) {
	document.getElementById("main").hidden = false;
	document.querySelector('body').removeChild(document.getElementById("login-form"));
	fetchData();
}