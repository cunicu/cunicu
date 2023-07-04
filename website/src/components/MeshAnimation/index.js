// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React, { useRef, useEffect } from "react";
import styles from './styles.module.css';

const opts = {
	particleColor: "rgb(120,120,120)",
	lineColor: "rgb(100,100,100)",
	particleAmount: 75,
	defaultSpeed: 0.3,
	variantSpeed: 1,
	defaultRadius: 3,
	variantRadius: 0,
	linkRadius: 200,
};

function Particle(x, y) {
	this.x = x;
	this.y = y;
	this.speed = opts.defaultSpeed + Math.random() * opts.variantSpeed;
	this.directionAngle = Math.floor(Math.random() * 360);
	this.color = opts.particleColor;
	this.radius = opts.defaultRadius + Math.random() * opts.variantRadius;

	this.vector = {
		x: Math.cos(this.directionAngle) * this.speed,
		y: Math.sin(this.directionAngle) * this.speed
	}

	this.update = (canvas) => {
		this.border(canvas);
		this.x += this.vector.x;
		this.y += this.vector.y;
	}

	this.border = (canvas) => {
		if (this.x >= canvas.width || this.x <= 0)
			this.vector.x *= -1;

		if (this.y >= canvas.height || this.y <= 0)
			this.vector.y *= -1;

		if (this.x > canvas.width)
			this.x = canvas.width;

		if (this.y > canvas.height)
			this.y = canvas.height;

		if (this.x < 0)
			this.x = 0;

		if (this.y < 0)
			this.y = 0;
	}

	this.draw = (context) => {
		context.beginPath();
		context.arc(this.x, this.y, this.radius, 0, Math.PI * 2);
		context.closePath();
		context.fillStyle = this.color;
		context.fill();
	};
};

function MeshAnimation() {
	var particles = [];
	var rgb = opts.lineColor.match(/\d+/g);

	let size;
	const canvasRef = useRef(null);
	const requestIdRef = useRef(null);


	const checkDistance = (x1, y1, x2, y2) => {
		return Math.sqrt(Math.pow(x2 - x1, 2) + Math.pow(y2 - y1, 2));
	}

	const linkPoints = (point1, hubs) => {
		let drawArea = canvasRef.current.getContext("2d");

		for (let i = 0; i < hubs.length; i++) {
			let distance = checkDistance(point1.x, point1.y, hubs[i].x, hubs[i].y);
			let opacity = 1 - distance / opts.linkRadius;
			if (opacity > 0) {
				drawArea.lineWidth = 0.5;
				drawArea.strokeStyle = `rgba(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, ${opacity})`;
				drawArea.beginPath();
				drawArea.moveTo(point1.x, point1.y);
				drawArea.lineTo(hubs[i].x, hubs[i].y);
				drawArea.closePath();
				drawArea.stroke();
			}
		}
	}

	const renderFrame = () => {
		let canvas = canvasRef.current;
		let context = canvas.getContext("2d");

		context.clearRect(0, 0, canvas.width, canvas.height);

		for (let i = 0; i < particles.length; i++) {
			particles[i].update(canvas);
			particles[i].draw(context);

			linkPoints(particles[i], particles);
		}
	};

	const tick = () => {
		if (!canvasRef.current) return;



		renderFrame();
		requestIdRef.current = requestAnimationFrame(tick);
	};

	useEffect(() => {
		let canvas = canvasRef.current;

		if (!canvas)
			return;

		const resizeObserver = new ResizeObserver(() => {
			canvas.height = canvas.clientHeight;
			canvas.width = canvas.clientWidth;

			if (canvas.height > 0 && canvas.width > 0 && particles.length == 0) {
				for (let i = 0; i < opts.particleAmount; i++) {
					let x = Math.random() * canvas.width;
					let y = Math.random() * canvas.height;

					particles.push(new Particle(x, y));
				}
			}
		});

		resizeObserver.observe(canvas);

		requestIdRef.current = requestAnimationFrame(tick);
		return () => {
			cancelAnimationFrame(requestIdRef.current);
			resizeObserver.disconnect();
		};
	}, []);

	return <canvas {...size} ref={canvasRef} className={styles.canvas} />;
}

export default MeshAnimation;