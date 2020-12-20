import {Component, OnInit} from '@angular/core';
import {User} from '@models/user';
import {UserService} from '@services/user.service';
import {finalize, first} from 'rxjs/operators';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';

@Component({templateUrl: './profile.component.html'})
export class ProfileComponent implements OnInit {
    initiated = false;
    loading = false;
    submitted = false;
    profile: User;
    profileForm: FormGroup;

    constructor(private formBuilder: FormBuilder, private userService: UserService) {
    }

    ngOnInit(): void {
        this.profileForm = this.formBuilder.group({
            username: ['', [Validators.required, Validators.minLength(4)]],
            firstName: ['', [Validators.maxLength(20)]],
            lastName: ['', [Validators.maxLength(30)]],
            age: ['', [Validators.min(18), Validators.max(120)]],
            gender: [''],
            city: ['', [Validators.maxLength(80)]],
            interests: ['', [Validators.maxLength(500)]],
        });

        this.userService.me()
            .pipe(finalize(() => this.initiated = true))
            .subscribe(p => this.profile = p);
    }

    // convenience getter for easy access to form fields
    get f() {
        return this.profileForm.controls;
    }

    onSubmit(): void {
        this.submitted = true;

        if (this.profileForm.invalid || this.loading) {
            return;
        }

        this.loading = true;
        this.userService.update(this.profileForm.value)
            .pipe(first(), finalize(() => this.loading = false))
            .subscribe(() => this.submitted = false);
    }
}
