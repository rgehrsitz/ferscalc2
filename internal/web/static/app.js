document.addEventListener('DOMContentLoaded', () => {
    // Set default dates
    const today = new Date();
    const thirtyYearsAgo = new Date(today.getFullYear() - 30, today.getMonth(), today.getDate());
    const twentyFiveYearsAgo = new Date(today.getFullYear() - 25, today.getMonth(), today.getDate());
    const fiveYearsFuture = new Date(today.getFullYear() + 5, today.getMonth(), today.getDate());

    document.getElementById('birthDate').valueAsDate = thirtyYearsAgo;
    document.getElementById('hireDate').valueAsDate = twentyFiveYearsAgo;
    document.getElementById('retireDate').valueAsDate = fiveYearsFuture;

    document.getElementById('calcForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());

        // Convert types
        const payload = {
            name: data.name,
            birthDate: data.birthDate,
            hireDate: data.hireDate,
            salary: parseFloat(data.salary),
            high3: parseFloat(data.high3),
            tspTrad: parseFloat(data.tspTrad),
            tspRoth: parseFloat(data.tspRoth),
            tspContrib: parseFloat(data.tspContrib) / 100,
            retireDate: data.retireDate,
            state: data.state,
            ssAge: parseInt(data.ssAge),
            ssEstimate: parseFloat(data.ssEstimate),
            tspStrategy: data.tspStrategy
        };

        try {
            const response = await fetch('/api/calculate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                throw new Error('Calculation failed');
            }

            const result = await response.json();
            displayResults(result);
        } catch (error) {
            console.error('Error:', error);
            alert('An error occurred during calculation.');
        }
    });
});

function displayResults(data) {
    const resultsDiv = document.getElementById('results');
    const contentDiv = document.getElementById('resultsContent');

    resultsDiv.classList.remove('hidden');
    contentDiv.innerHTML = '';

    // Helper to format currency
    const fmt = (num) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(num);

    // Create summary
    const summary = document.createElement('div');
    summary.innerHTML = `
        <div class="result-item">
            <span>Retirement Date</span>
            <span class="result-value">${data.retirementDate}</span>
        </div>
        <div class="result-item">
            <span>FERS Pension (Annual)</span>
            <span class="result-value">${fmt(data.fersAnnual)}</span>
        </div>
        <div class="result-item">
            <span>Social Security Supplement (Annual)</span>
            <span class="result-value">${fmt(data.supplementAnnual)}</span>
        </div>
        <div class="result-item">
            <span>Projected TSP Balance</span>
            <span class="result-value">${fmt(data.tspBalance)}</span>
        </div>
        <div class="result-item">
            <span>First Year Net Income</span>
            <span class="result-value">${fmt(data.netIncome)}</span>
        </div>
        <div class="result-item">
            <span>First Year TSP Withdrawal</span>
            <span class="result-value">${fmt(data.tspWithdrawal)}</span>
        </div>
        <div class="result-item">
            <span>TSP Strategy</span>
            <span class="result-value">${data.strategyDescription}</span>
        </div>
    `;

    contentDiv.appendChild(summary);
}
