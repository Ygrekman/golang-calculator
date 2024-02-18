// Отправить вычисление на сервер
function submitCalculation() {
    const expression = document.getElementById('expression').value; // Получаем выражение от пользователя
    const calculationResultsSection = document.getElementById('calculation-results'); // Получаем секцию для вывода результатов

    // Проверяем выражение на валидность и на деление на ноль
    if (/\/0/.test(expression)) {
        appendCalculationResult(calculationResultsSection, null, `[${expression}] - Division by zero is not allowed.`, 'error');
        return;
    }
    // Улучшенное регулярное выражение для проверки валидности арифметического выражения
    if (!/^\d+([\+\-\*\/]\(?\-?\d+\)?)+$/.test(expression)) {
        appendCalculationResult(calculationResultsSection, null, `[${expression}] - Invalid expression format.`, 'error');
        return;
    }

    // Отправляем запрос на сервер
    fetch('http://localhost:8080/submit-calculation', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
            operation: expression,
            add_duration: parseInt(document.getElementById('plus-time').value),
            subtract_duration: parseInt(document.getElementById('minus-time').value),
            multiply_duration: parseInt(document.getElementById('multiply-time').value),
            divide_duration: parseInt(document.getElementById('divide-time').value),
            inactive_server_time: parseInt(document.getElementById('inactive-server-time').value),
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'created') {
             // Сохраняем ID в localStorage
            let successfulIDs = JSON.parse(localStorage.getItem('successfulIDs')) || [];
            successfulIDs.push(data.id);
            localStorage.setItem('successfulIDs', JSON.stringify(successfulIDs));
            
            // Добавляем новый элемент для отображения операции и статуса
            appendCalculationResult(calculationResultsSection, data.id, data.operation, 'pending');
        } else if (data.status === 'error') {
            // Добавляем сообщение об ошибке и операцию
            appendCalculationResult(calculationResultsSection, `${expression} - ${data.error}`, 'error');
        }
    })
    .catch((error) => {
        console.error('Error:', error);
        appendCalculationResult(calculationResultsSection, `${expression} - Error: ${error}`, 'error');
    });
}

// Функция для загрузки всех вычислений и предотвращения дубликатов
function loadAllCalculations() {
    // Очищаем существующие вычисления, чтобы избежать дублирования
    const calculationResultsSection = document.getElementById('calculation-results');
    calculationResultsSection.innerHTML = '';

    fetch('http://localhost:8080/get-all-calculations')
        .then(response => response.json())
        .then(data => {
            data.forEach(calculation => {
                const status = calculation.status === 'completed' ? 'success' : 'pending';
                const resultText = calculation.status === 'completed' ? calculation.result : '?';
                appendCalculationResult(calculationResultsSection, calculation.id, `${calculation.operation} Result = ${resultText}`, status);
            });
        })
        .catch(error => console.error('Error loading calculations:', error));
}

// Функция appendCalculationResult для динамического контента в зависимости от статуса
function appendCalculationResult(parentElement, id, message, status) {
    const resultElement = document.createElement('div');
    resultElement.className = `calculation-result ${status}`;
    resultElement.id = `result-${id}`;

    const idLine = document.createElement('div');
    idLine.textContent = `ID: ${id}`;
    resultElement.appendChild(idLine);

    const operationLine = document.createElement('div');
    operationLine.textContent = message;
    resultElement.appendChild(operationLine);

    if (status === 'pending') {
        const pendingLine = document.createElement('div');
        pendingLine.textContent = 'Expression will be calculated soon.';
        resultElement.appendChild(pendingLine);
    }

    parentElement.appendChild(resultElement);
}

// Сохранение настроек (в этом примере используется локальное хранилище)
function saveSettings() {
    localStorage.setItem('plus-time', document.getElementById('plus-time').value);
    localStorage.setItem('minus-time', document.getElementById('minus-time').value);
    localStorage.setItem('multiply-time', document.getElementById('multiply-time').value);
    localStorage.setItem('divide-time', document.getElementById('divide-time').value);
    localStorage.setItem('inactive-server-time', document.getElementById('inactive-server-time').value);
    
    alert('Settings saved successfully.');
}

// Функция для получения статуса серверов
function fetchServerStatuses() {
    const serverStatusesDiv = document.getElementById('server-statuses');
    // Очистка существующих статусов для исключения дубликатов
    serverStatusesDiv.innerHTML = '';

    // Получение статуса сервера orchestrator
    fetch('http://localhost:8080/orchestrator-status')
    .then(response => response.json())
    .then(orchestratorStatus => {
        // Отображение статус сервера orchestrator
        const orchestratorDiv = document.createElement('div');
        orchestratorDiv.className = orchestratorStatus.running ? 'server-status running' : 'server-status error';
        orchestratorDiv.innerHTML = `
            <p><strong>Type:</strong> Orchestrator</p>
            <p><strong>URL:</strong> http://localhost:8080/</p>
            <p><strong>Status:</strong> ${orchestratorStatus.running ? 'Running' : 'Not Running'} - ${orchestratorStatus.message}</p>
        `;
        serverStatusesDiv.appendChild(orchestratorDiv);
    })
    .catch(error => {
        console.error('Error fetching orchestrator status:', error);
        appendErrorServerDiv(serverStatusesDiv, 'Orchestrator', 'http://localhost:8080/', 'Unavailable - Could not connect');
    });

    // Получение статусов серверов calculator
    fetch('http://localhost:8080/ping-servers')
    .then(response => response.json())
    .then(calculatorStatuses => {
        calculatorStatuses.forEach(server => {
            appendServerStatusDiv(serverStatusesDiv, server);
        });
    })
    .catch(error => {
        console.error('Error fetching calculator servers statuses:', error);
        appendErrorServerDiv(serverStatusesDiv, 'Calculator', 'Unavailable URL', 'Unavailable - Could not connect');
    });
}

// Функция для добавления div элементов статусов серверов
function appendServerStatusDiv(parentElement, server) {
    const serverDiv = document.createElement('div');
    serverDiv.className = server.running ? 'server-status running' : 'server-status error';
    serverDiv.innerHTML = `
        <p><strong>Type:</strong> Calculator</p>
        <p><strong>URL:</strong> ${server.url}</p>
        <p><strong>Status:</strong> ${server.running ? 'Running' : 'Not Running'}${server.error ? ` (Error: ${server.error})` : ''}</p>
        ${server.running ? `<p><strong>Max Goroutines:</strong> ${server.maxGoroutines}</p>
        <p><strong>Current Goroutines:</strong> ${server.currentGoroutines}</p>` : ''}
    `;
    parentElement.appendChild(serverDiv);
}

// Функция для добавления div элементов ошибочных статусов серверов
function appendErrorServerDiv(parentElement, type, url, status) {
    const errorDiv = document.createElement('div');
    errorDiv.className = 'server-status error';
    errorDiv.innerHTML = `
        <p><strong>Type:</strong> ${type}</p>
        <p><strong>URL:</strong> ${url}</p>
        <p><strong>Status:</strong> ${status}</p>
    `;
    parentElement.appendChild(errorDiv);
}

// Функция для обновления результатов операций
function updateResults() {
    // Получение всех 'pending' операций на странице
    const pendingResults = document.querySelectorAll('.calculation-result.pending');

    pendingResults.forEach(resultElement => {
        const id = resultElement.id.split('-')[1]; // Предполагается, что формат ID - "result-{id}"

        // Запрашиваем результат операции по ID
        fetch(`http://localhost:8080/get-calculation-result?id=${id}`)
            .then(response => response.json())
            .then(data => {
                if (data.status === 'completed' && data.result !== undefined) {
                    // Обновляем текст результата и класс элемента
                    const operationLine = resultElement.querySelector('div:last-child');
                    operationLine.textContent = `[${data.operation}] Result = ${data.result}`;
                    resultElement.classList.remove('pending');
                    resultElement.classList.add('success');
                    resultElement.style.backgroundColor = "#4CAF50"; // Зеленый фон для завершенных операций
                } else {
                    // Если статус не завершен или результат отсутствует, оставляем как есть
                    console.log(`Calculation ID ${id} is still pending.`);
                }
            })
            .catch(error => console.error('Error updating result:', error));
    });
}

// Функция для очистки и обновления результатов операций
function clearAllCalculationsAndUpdate() {
    fetch('http://localhost:8080/clear-all-calculations', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (response.ok) {
            console.log('All calculations cleared successfully.');
            // Очищаем отображаемые результаты
            document.getElementById('calculation-results').innerHTML = '';
            // По желанию, повторно загружаем и отображаем все вычисления
            updateResults();
        } else {
            console.error('Failed to clear calculations.');
        }
    })
    .catch(error => console.error('Error:', error));
}

document.addEventListener('DOMContentLoaded', function() {
    // Начальная настройка и обработчики событий
    fetchServerStatuses();
    document.getElementById('reload-server-status').addEventListener('click', fetchServerStatuses);
    document.getElementById('reload-operations-status').addEventListener('click', updateResults);
    loadAllCalculations(); // Загрузка всех вычислений при загрузке страницы

    // Автоматическое обновление всех вычислений каждую минуту
    setInterval(updateResults, 60000);

    // Применение настроек изначально и каждый раз при их сохранении
    applySettings();
    document.querySelector('button[onclick="saveSettings()"]').addEventListener('click', applySettings);
});

// Адаптация функции применения настроек и обработки блоков с ошибками сервера
function applySettings() {
    // Получение и применение настройки времени неактивности сервера
    const inactiveServerTime = parseInt(localStorage.getItem('inactive-server-time'), 10) || 60; // Значение по умолчанию 60 секунд
    localStorage.setItem('inactive-server-time', inactiveServerTime); // Обновление, если использовалось значение по умолчанию
    console.log(`Settings applied. Inactive server time: ${inactiveServerTime} seconds.`);

    // Настройка таймаута для очистки блоков с ошибками на основе времени неактивности сервера
    setTimeout(() => {
        const errorDivs = document.querySelectorAll('.server-status.error');
        errorDivs.forEach(div => div.remove());
        console.log('Old error divs removed based on inactive server time setting.');
    }, inactiveServerTime * 1000); // Перевод секунд в миллисекунды
}