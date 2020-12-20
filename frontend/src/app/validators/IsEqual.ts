import {FormGroup} from '@angular/forms';

// Custom validator to check that two fields equals
// tslint:disable-next-line:typedef
export function IsEqual(controlName: string, matchingControlName: string) {
    return (formGroup: FormGroup) => {
        const control = formGroup.controls[controlName];
        const matchingControl = formGroup.controls[matchingControlName];

        // Return if another validator has already found an error on the matchingControl
        if (matchingControl.errors && !matchingControl.errors.mustMatch) return;

        // Set error on matchingControl if validation fails
        matchingControl.setErrors(control.value !== matchingControl.value ? {isEqual: true} : null);
    };
}
