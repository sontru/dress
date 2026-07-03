document.addEventListener('DOMContentLoaded', () => {
    setupHeaderAuth();
});

async function setupHeaderAuth() {
    const authButton = document.getElementById('authButton');
    const myMediaButton = document.getElementById('myMediaButton');
    const registerButtons = document.querySelectorAll('[data-auth-register], #registerButton');
    if (!authButton && !registerButtons.length) return;

    try {
        const response = await fetch(appPath('/api/me'));
        if (!response.ok) {
            if (authButton) {
                authButton.textContent = 'Sign in';
                authButton.href = appPath('/login');
            }
            if (myMediaButton) myMediaButton.style.display = 'none';
            registerButtons.forEach((button) => {
                button.style.display = '';
            });
            return;
        }

        if (authButton) {
            authButton.textContent = 'Sign out';
            authButton.href = '#';
        }
        if (myMediaButton) myMediaButton.style.display = '';
        registerButtons.forEach((button) => {
            button.style.display = 'none';
        });

        authButton?.addEventListener('click', async (event) => {
            event.preventDefault();
            authButton.style.pointerEvents = 'none';

            try {
                const logoutResponse = await fetch(appPath('/api/logout'), { method: 'POST' });
                if (!logoutResponse.ok) throw new Error(await logoutResponse.text());
                window.location.href = appPath('/');
            } catch (error) {
                console.error('Error signing out:', error);
                authButton.style.pointerEvents = '';
            }
        });
    } catch (error) {
        if (authButton) {
            authButton.textContent = 'Sign in';
            authButton.href = appPath('/login');
        }
        if (myMediaButton) myMediaButton.style.display = 'none';
        registerButtons.forEach((button) => {
            button.style.display = '';
        });
    }
}
