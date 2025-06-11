
const ACCOUNT_ID="e676b46d-45bd-4a20-8529-f5b47ae35d07"
const BASE_URL = "http://localhost:8080"
// set access control allow origin
const corsHeaders = {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type, X-Authorized-User"
};
// set headers
const headers = new Headers(corsHeaders);

// On page load get the use balance and transactions
document.addEventListener("DOMContentLoaded", function () {
    // Check if the user is logged in
    const isLoggedIn = localStorage.getItem("isLoggedIn");
    const authPrompt = document.getElementById("authPrompt");
    const mainContent = document.getElementById("mainContent");
    // Try a quick auth check if token exists
    const token = localStorage.getItem("token");
    if (token) {
        // Optionally, ping a protected endpoint to check validity
        fetch(`${BASE_URL}/api/${localStorage.getItem("accountId")}/details`, {
            headers: { "Authorization": token }
        })
        .then(res => {
            if (res.status === 401) {
                // Token is invalid, clear local storage
                console.log("Getting a 401!!")
                localStorage.clear();
                if (authPrompt) authPrompt.style.display = "block";
                if (mainContent) authPrompt.style.display = "none";
                return;
            }
            if (!isLoggedIn) {
                if (authPrompt) authPrompt.style.display = "block";
                if (mainContent) mainContent.style.display = "none";
                return;
            } else {
                if (authPrompt) authPrompt.style.display = "none";
                if (mainContent) mainContent.style.display = "block";
            }
            const accountId = localStorage.getItem("accountId");
            if (accountId) {
                getAccountDetails(accountId);
                getAccountTransactions(accountId);
            } else {
                console.error("No account ID found in local storage.");
            }
        })
        .catch(() => {
            localStorage.clear();
            if (authPrompt) authPrompt.style.display = "block";
            if (mainContent) mainContent.style.display = "none";
        });
        return;
    }
    if (!isLoggedIn) {
        if (authPrompt) authPrompt.style.display = "block";
        if (mainContent) mainContent.style.display = "none";
        return;
    } else {
        if (authPrompt) authPrompt.style.display = "none";
        if (mainContent) mainContent.style.display = "block";
    }

    const accountId = localStorage.getItem("accountId");
    if (accountId) {
        getAccountDetails(accountId);
        getAccountTransactions(accountId);
    } else {
        console.error("No account ID found in local storage.");
    }

});