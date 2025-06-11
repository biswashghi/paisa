// Function to get account details
function getAccountDetails(accountId) {

    // fetch account details
    fetch(`${BASE_URL}/api/${accountId}/details`, {
        method: "GET",
        headers: {
            ...corsHeaders,
            "Authorization": localStorage.getItem("token")
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error("Network response was not ok");
            }
            return response.json();
        })
        .then(data => {
            // data format: {accountId: 'e676b46d-45bd-4a20-8529-f5b47ae35d07', rewardsBalance: '90'}
            console.log("Account details:", data);
            const balanceElement = document.getElementById("balance");
            balanceElement.textContent = `Balance: ${data.rewardsBalance}`;
        })
        .catch(error => {
            console.error("Error fetching account details:", error);
        });
}

function refreshBalance() {
    // const accountId = localStorage.getItem("accountId");
    const accountId = ACCOUNT_ID;
    if (accountId) {
        getAccountDetails(accountId);
    } else {
        console.error("No account ID found in local storage.");
    }
}

let currentPage = 0;

// Function to get account transactions
function getAccountTransactions(accountId, page = 0, pageSize = 10) {
    currentPage = page;
    fetch(`${BASE_URL}/api/list/transactions/${accountId}?start_page=${page}&page_size=${pageSize}`, {
        method: "GET",
        headers: {
            ...corsHeaders,
            "Authorization": localStorage.getItem("token")
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error("Network response was not ok");
            }
            return response.json();
        })
        .then(data => {
            // Example response data
            // [{accountId:..., description:..., merchantCode:..., increment:..., createdAt:..., transactionId:...}, ...]
            console.log("Account transactions:", data);
            const transactionsBody = document.getElementById("transactionsBody");
            if (!transactionsBody) {
                console.error("transactionsBody element not found in DOM.");
                return;
            }
            transactionsBody.innerHTML = "";
            data.forEach(transaction => {
                const row = document.createElement("tr");
                row.innerHTML = `
                    <td>${transaction.transactionId}</td>
                    <td>${transaction.description}</td>
                    <td>${transaction.merchantCode}</td>
                    <td>${transaction.increment}</td>
                    <td>${new Date(transaction.createdAt).toLocaleString()}</td>
                `;
                transactionsBody.appendChild(row);
            });
            // Set up next page button
            const nextPageBtn = document.getElementById("nextPageBtn");
            if (nextPageBtn) {
                nextPageBtn.onclick = function() {
                    getAccountTransactions(accountId, currentPage + 1, pageSize);
                };
            }
        })
        .catch(error => {
            console.error("Error fetching account transactions:", error);
        });
}

function refreshTransactions() {
    // const accountId = localStorage.getItem("accountId");
    const accountId = ACCOUNT_ID;
    if (accountId) {
        getAccountTransactions(accountId);
    } else {
        console.error("No account ID found in local storage.");
    }
}
