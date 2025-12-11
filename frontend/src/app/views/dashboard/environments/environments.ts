import { Component, inject, OnInit, signal, WritableSignal } from '@angular/core';
import { Environment, EnvironmentData } from '../../../services/environment/environment';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-environments',
  imports: [CommonModule],
  templateUrl: './environments.html',
  styleUrl: './environments.css',
})
export class Environments implements OnInit {
  private environmentService = inject(Environment);
  
  environments$: WritableSignal<EnvironmentData[]> = signal([]);
  
  ngOnInit(): void {
    this.environmentService.getAll().subscribe({
      next: (data) => this.environments$.set(data)
    })
  }
}
