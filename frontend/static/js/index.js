// Dynamically load HTML partials
document.addEventListener("DOMContentLoaded", function() {
    fetch('wallet_balance.html').then(r => r.text()).then(html => {
        document.getElementById('walletBalanceContainer').innerHTML = html;
    });
    fetch('transactions.html').then(r => r.text()).then(html => {
        document.getElementById('transactionsContainer').innerHTML = html;
        // Now it's safe to call getAccountTransactions
        const accountId = localStorage.getItem("accountId");
        if (accountId) {
            getAccountTransactions(accountId);
        }
    });
});