document.addEventListener('DOMContentLoaded', function() {
    const orderIdInput = document.getElementById('orderId');
    const searchBtn = document.getElementById('searchBtn');
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    const responseTime = document.getElementById('responseTime');
    const orderData = document.getElementById('orderData');

    searchBtn.addEventListener('click', searchOrder);
    orderIdInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            searchOrder();
        }
    });

    function searchOrder() {
        const orderId = orderIdInput.value.trim();
        if (!orderId) {
            showError('Please enter an Order ID');
            return;
        }

        // Reset UI
        hideError();
        hideOrderData();
        showLoading();

        // Start timing
        const startTime = performance.now();

        fetch(`http://localhost:8000/api/order/${orderId}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(response.status === 404 ? 'Order not found' : 'Error fetching order');
                }
                return response.json();
            })
            .then(data => {
                // Calculate response time
                const endTime = performance.now();
                const timeTaken = (endTime - startTime).toFixed(2);

                displayOrder(data, timeTaken);
            })
            .catch(err => {
                showError(err.message);
            })
            .finally(() => {
                hideLoading();
            });
    }

    function displayOrder(data, timeTaken) {
        responseTime.textContent = `Found in ${timeTaken} ms`;
        responseTime.classList.remove('hidden');

        orderData.innerHTML = ''; // clear previous
        renderTreeJSON(data, orderData);
        orderData.classList.remove('hidden');
    }

    function showLoading() {
        loading.classList.remove('hidden');
    }

    function hideLoading() {
        loading.classList.add('hidden');
    }

    function showError(message) {
        error.textContent = message;
        error.classList.remove('hidden');
    }

    function hideError() {
        error.classList.add('hidden');
    }

    function hideOrderData() {
        orderData.classList.add('hidden');
        responseTime.classList.add('hidden');
    }
});

function renderTreeJSON(obj, container) {
    const root = createTree(obj);
    container.appendChild(root);
}

function createTree(value, key = null) {
    const container = document.createElement('div');
    container.classList.add('json-container');

    const isObject = typeof value === 'object' && value !== null;
    const isArray = Array.isArray(value);

    if (isObject) {
        const wrapper = document.createElement('div');
        wrapper.classList.add('json-node');

        const toggle = document.createElement('span');
        toggle.classList.add('json-toggle');
        toggle.textContent = '▾'; // down arrow
        toggle.addEventListener('click', () => {
            wrapper.classList.toggle('collapsed');
            toggle.textContent = wrapper.classList.contains('collapsed') ? '▸' : '▾';
        });

        const keySpan = document.createElement('span');
        keySpan.classList.add('json-key');
        keySpan.textContent = key !== null ? `"${key}": ` : '';

        const typeLabel = document.createElement('span');
        typeLabel.textContent = isArray ? '[...]' : '{...}';

        const childrenContainer = document.createElement('div');
        childrenContainer.classList.add('json-children');

        for (const k in value) {
            const child = createTree(value[k], k);
            childrenContainer.appendChild(child);
        }

        wrapper.appendChild(toggle);
        wrapper.appendChild(keySpan);
        wrapper.appendChild(typeLabel);
        wrapper.appendChild(childrenContainer);

        container.appendChild(wrapper);
    } else {
        const row = document.createElement('div');
        if (key !== null) {
            const keySpan = document.createElement('span');
            keySpan.classList.add('json-key');
            keySpan.textContent = `"${key}": `;
            row.appendChild(keySpan);
        }

        const valueSpan = document.createElement('span');
        valueSpan.classList.add(getValueClass(value));
        valueSpan.textContent = formatValue(value);

        row.appendChild(valueSpan);
        container.appendChild(row);
    }

    return container;
}


function getValueClass(value) {
    if (typeof value === 'string') return 'json-string';
    if (typeof value === 'number') return 'json-number';
    if (typeof value === 'boolean') return 'json-boolean';
    if (value === null) return 'json-null';
    return '';
}

function formatValue(value) {
    if (typeof value === 'string') return `"${value}"`;
    if (value === null) return 'null';
    return value;
}