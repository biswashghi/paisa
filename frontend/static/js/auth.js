//Login handler
function login(event) {
    event.preventDefault();
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    // Perform login request
    fetch(`${BASE_URL}/login`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            ...corsHeaders
        },
        body: JSON.stringify({ username, password })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error("Login failed");
            }
            return response.json();
        })
        .then(data => {
            // Store the sessionId and accountId in local storage
            localStorage.setItem("token", data.sessionId);
            localStorage.setItem("accountId", data.accountId);
            localStorage.setItem("isLoggedIn", true);
            window.location.href = "index.html"; // Redirect to main page
        })
        .catch(error => {
            console.error("Error:", error);
            alert("Login failed. Please check your credentials.");
            window.location.href = "index.html";
        });
}

// Register handler
function register(event) {
    event.preventDefault();
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    fetch(`${BASE_URL}/register`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            ...corsHeaders
        },
        body: JSON.stringify({ username, password })
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => { throw new Error(data.error || "Registration failed"); });
            }
            return response.json();
        })
        .then(data => {
            alert("Registration successful! Please login.");
            window.location.href = "login.html";
        })
        .catch(error => {
            console.error("Error:", error);
            alert(error.message || "Registration failed. Please try again.");
        });
}