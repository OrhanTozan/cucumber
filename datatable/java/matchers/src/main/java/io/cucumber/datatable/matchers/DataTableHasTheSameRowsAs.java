package io.cucumber.datatable.matchers;

import io.cucumber.datatable.DataTable;
import io.cucumber.datatable.DataTableDiff;
import io.cucumber.datatable.TableDiffer;
import org.hamcrest.Description;
import org.hamcrest.Factory;
import org.hamcrest.TypeSafeDiagnosingMatcher;

public class DataTableHasTheSameRowsAs extends TypeSafeDiagnosingMatcher<DataTable> {
    private final DataTable expectedValue;
    private final boolean unordered;

    private DataTableHasTheSameRowsAs(DataTable expectedValue, boolean unordered) {
        this.expectedValue = expectedValue;
        this.unordered = unordered;
    }

    @Override
    public void describeTo(Description description) {
        description.appendText("a datable with the same rows");
        if (unordered) {
            description.appendText(" in any order");
        }
    }


    @Override
    protected boolean matchesSafely(DataTable item, Description description) {
        TableDiffer tableDiffer = new TableDiffer(expectedValue, item);
        DataTableDiff diff = unordered ? tableDiffer.calculateUnorderedDiffs() : tableDiffer.calculateDiffs();

        if (diff.isEmpty()) {
            return true;
        }
        description.appendText("the tables were different\n");
        description.appendText(diff.toString());
        return false;
    }

    @Factory
    public static DataTableHasTheSameRowsAs hasTheSameRowsInOrderAs(DataTable operand) {
        return new DataTableHasTheSameRowsAs(operand, true);
    }

    @Factory
    public static DataTableHasTheSameRowsAs hasTheSameRowsAs(DataTable operand) {
        return new DataTableHasTheSameRowsAs(operand, false);
    }


}
