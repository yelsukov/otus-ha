import {Component, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';

import {AuthService} from '@services/auth.service';
import {IsEqual} from '../../validators/IsEqual';
import {finalize, first} from 'rxjs/operators';

@Component({templateUrl: './registration.component.html'})
export class RegistrationComponent implements OnInit {
    registerForm: FormGroup;
    submitted = false;
    loading = false;
    success = false;

    constructor(private formBuilder: FormBuilder, private authService: AuthService) {
    }

    ngOnInit(): void {
        this.registerForm = this.formBuilder.group({
                username: ['', [Validators.required, Validators.minLength(4)]],
                password: ['', [Validators.required, Validators.minLength(6)]],
                passwordConfirm: ['', [Validators.required]],

                firstName: ['', [Validators.maxLength(20)]],
                lastName: ['', [Validators.maxLength(30)]],
                age: ['', [Validators.min(18), Validators.max(120)]],
                gender: [''],
                city: ['', [Validators.maxLength(80)]],
            },
            {validator: IsEqual('password', 'passwordConfirm')});
    }

    // convenience getter for easy access to form fields
    get f() {
        return this.registerForm.controls;
    }

    onSubmit(): void {
        this.submitted = true;

        if (this.registerForm.invalid || this.loading) {
            return;
        }

        this.loading = true;
        this.authService.signUp(this.registerForm.value)
            .pipe(first(), finalize(() => this.loading = false))
            .subscribe(() => this.success = true);
    }
}
