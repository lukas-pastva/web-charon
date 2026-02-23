// Web Charon - Motoklub Charon

document.addEventListener('DOMContentLoaded', function () {
    // Mobile nav toggle
    const navToggle = document.querySelector('.nav-toggle');
    const navLinks = document.querySelector('.nav-links');
    if (navToggle && navLinks) {
        navToggle.addEventListener('click', function () {
            navLinks.classList.toggle('active');
        });
    }

    // Theme switcher
    initThemeSwitcher();

    // Gallery lightbox
    initLightbox();
});

// Theme switcher with localStorage persistence
function initThemeSwitcher() {
    var themes = ['flame', 'red', 'steel'];
    var btn = document.getElementById('themeToggle');
    if (!btn) return;

    // Load saved theme
    var saved = localStorage.getItem('charon-theme');
    if (saved && themes.indexOf(saved) !== -1) {
        document.documentElement.setAttribute('data-theme', saved);
    }

    btn.addEventListener('click', function () {
        var current = document.documentElement.getAttribute('data-theme') || 'flame';
        var idx = themes.indexOf(current);
        var next = themes[(idx + 1) % themes.length];
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('charon-theme', next);
    });
}

function initLightbox() {
    const images = document.querySelectorAll('.gallery-images img');
    if (images.length === 0) return;

    // Create lightbox element
    const lightbox = document.createElement('div');
    lightbox.className = 'lightbox';
    lightbox.innerHTML = `
        <button class="lightbox-close" aria-label="Zavrieť">&times;</button>
        <button class="lightbox-nav lightbox-prev" aria-label="Predchádzajúci">&lsaquo;</button>
        <button class="lightbox-nav lightbox-next" aria-label="Ďalší">&rsaquo;</button>
        <img src="" alt="">
        <div class="lightbox-caption"></div>
    `;
    document.body.appendChild(lightbox);

    const lbImg = lightbox.querySelector('img');
    const lbCaption = lightbox.querySelector('.lightbox-caption');
    const btnClose = lightbox.querySelector('.lightbox-close');
    const btnPrev = lightbox.querySelector('.lightbox-prev');
    const btnNext = lightbox.querySelector('.lightbox-next');

    let currentIndex = 0;

    function showImage(index) {
        if (index < 0) index = images.length - 1;
        if (index >= images.length) index = 0;
        currentIndex = index;
        lbImg.src = images[index].src;
        lbCaption.textContent = images[index].getAttribute('data-caption') || '';
    }

    function openLightbox(index) {
        showImage(index);
        lightbox.classList.add('active');
        document.body.style.overflow = 'hidden';
    }

    function closeLightbox() {
        lightbox.classList.remove('active');
        document.body.style.overflow = '';
    }

    images.forEach(function (img, i) {
        img.addEventListener('click', function () {
            openLightbox(i);
        });
    });

    btnClose.addEventListener('click', closeLightbox);
    btnPrev.addEventListener('click', function () { showImage(currentIndex - 1); });
    btnNext.addEventListener('click', function () { showImage(currentIndex + 1); });

    lightbox.addEventListener('click', function (e) {
        if (e.target === lightbox) closeLightbox();
    });

    document.addEventListener('keydown', function (e) {
        if (!lightbox.classList.contains('active')) return;
        if (e.key === 'Escape') closeLightbox();
        if (e.key === 'ArrowLeft') showImage(currentIndex - 1);
        if (e.key === 'ArrowRight') showImage(currentIndex + 1);
    });
}

// Confirm delete actions
document.addEventListener('click', function (e) {
    if (e.target.closest('[data-confirm]')) {
        var msg = e.target.closest('[data-confirm]').getAttribute('data-confirm');
        if (!confirm(msg)) {
            e.preventDefault();
        }
    }
});
