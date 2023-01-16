import { useForm, Controller } from 'react-hook-form';
import TextField from '@mui/material/TextField';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import InputAdornment from '@mui/material/InputAdornment';
import { InputProps } from '@mui/material';
import { styled } from '@mui/material/styles';
import Tooltip, { tooltipClasses, TooltipProps } from '@mui/material/Tooltip';
import { grey } from '@mui/material/colors';

export enum FieldType {
    VALUE_BASED = 'value-based-field',
    LIST_BASED = 'list-based-field',
    OBJECT_BASED = 'object-based-field',
    SINGLE_SELECTION_BASED = 'single-selection-field',
    TIME_BASED = 'time-based-field' // append 's' to the output representing duration in  seconds
};

export type FormField = {
    name: string;
    label: string;
    default?: string | number;
    info?: string;
    fieldType: FieldType,
    keys?: string[],
    required?: string | boolean;
    options?: { name: string, value: string|number, info?: string }[],
};

interface Props {
    handleFormSubmission: Function;
    fields: FormField[];
    submitRef: React.Ref<HTMLButtonElement>;
};


const HTMLTooltip = styled(({ className, ...props }: TooltipProps) => (
    <Tooltip {...props} classes={{ popper: className }} />
))(( { theme }) => ({
    [`& .${tooltipClasses.tooltip}`]: {
        backgroundColor: grey[100],
        color: 'rgba(0, 0, 0, 0.80)',
        maxWidth: 220,
        fontSize: theme.typography.pxToRem(12),
        border: '1px solid #dadde9' // an arbitrary color darker then bgcolor
    }
}));

export default function Inputs(props: Props) {
    const {
        handleFormSubmission,
        fields,
        submitRef,
    } = props;

    const defaultValues: { [key: string]: any } = fields
        /**
         * Cannot apply undefined to defaultValue or defaultValues at useForm
         * Keeping a default value as undefined would cause an error from React
         * about turning an uncontrolled input to a controlled input
         * @see https://react-hook-form.com/api/usecontroller/controller 
         */
        .reduce((prev, next) => ({
            ...prev,
            [next.name]: next.default || '',
        }), {});

    const { control, handleSubmit } = useForm({ defaultValues });

    const onSubmit = (data: any) => {
        handleFormSubmission?.(data);
    };

    const getEndInputProps = (fieldLabel: string = '', fieldInfo: string = ''): InputProps => {
        return {
            endAdornment: (
                <InputAdornment position='end'>
                    <HTMLTooltip
                        placement='bottom-end'
                        title={
                            <>
                                <Typography color='inherit' variant='body2' fontWeight='bold' gutterBottom>
                                    {fieldLabel} &nbsp;
                                </Typography>
                                <Typography variant='caption'>
                                    {fieldInfo}
                                </Typography>
                            </>
                        }
                    >
                        <HelpOutlineIcon fontSize='small' />
                    </HTMLTooltip>
                </InputAdornment>
            )
        }
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Grid container spacing={2} sx={{ mb: 3 }}>
                {fields.map((f) => {
                    return (
                        <Grid item xs={12} md={3} key={f.name}>
                            <Controller
                                name={f.name}
                                control={control}
                                rules={{ required: f.required }}
                                render={({ field, fieldState: { error } }) => {
                                    if (f.fieldType === FieldType.TIME_BASED) {
                                        return (
                                            <TextField
                                                fullWidth
                                                variant='outlined'
                                                label={f.label}
                                                size='small'
                                                type='number'
                                                error={error !== undefined}
                                                InputProps={{ ...getEndInputProps(f.label, f.info) }}
                                                helperText={error !== undefined ? error.message : ''}
                                                {...field}
                                            />
                                        );
                                    }
                                    if (f.fieldType === FieldType.SINGLE_SELECTION_BASED) {
                                        return (
                                            <TextField
                                                fullWidth
                                                select={true}
                                                variant='outlined'
                                                label={f.label}
                                                size='small'
                                                error={error !== undefined}
                                                InputProps={{ ...getEndInputProps(f.label, f.info) }}
                                                helperText={error !== undefined ? error.message : ''}
                                                {...field}
                                            >
                                                {f.options?.map((p) =>
                                                    <MenuItem value={p.value} key={p.value}>
                                                        <Typography variant='caption'>
                                                            {p.name}
                                                        </Typography>
                                                    </MenuItem>
                                                )}
                                            </TextField>
                                        );
                                    }
                                    return (
                                        <TextField
                                            fullWidth
                                            variant='outlined'
                                            label={f.label}
                                            size='small'
                                            error={error !== undefined}
                                            InputProps={{ ...getEndInputProps(f.label, f.info) }}
                                            helperText={error !== undefined ? error.message : ''}
                                            {...field}
                                        />
                                    );
                                }}
                            />
                        </Grid>
                    );
                })}
            </Grid>
            <Button
                type="submit"
                ref={submitRef}
                sx={{ display: 'none' }}
            />
        </form>
    );
}
